// Package network implements hexagonal mesh networking for the Hexagonal Chain
package network

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"

	hexcore "github.com/hexagonal-chain/hexchain/pkg/core"
)

const (
	// Protocol constants
	HexMeshProtocolName    = "hexmesh"
	HexMeshProtocolVersion = 1
	HexMeshProtocolLength  = 20

	// Message codes
	HexBlockMsg         = 0x10
	HexHeaderMsg        = 0x11
	HexBlockRequestMsg  = 0x12
	HexHeaderRequestMsg = 0x13
	HexProofMsg         = 0x14
	HexStatusMsg        = 0x15
	HexNeighborMsg      = 0x16
	HexMeshStateMsg     = 0x17

	// Network constants
	MaxNeighborPeers      = 6   // Maximum neighbors in hex topology
	MaxConcurrentRequests = 100 // Maximum concurrent requests
	RequestTimeout        = 30  // Seconds
	HeartbeatInterval     = 15  // Seconds
)

// HexMeshProtocol implements the hexagonal mesh networking protocol
type HexMeshProtocol struct {
	config  *HexMeshConfig
	peers   map[enode.ID]*HexPeer
	peersMu sync.RWMutex

	// Network state
	localPosition hexcore.HexCoordinate
	networkID     uint64
	currentHead   common.Hash

	// Communication channels
	blockCh  chan *hexcore.HexBlock
	headerCh chan *hexcore.HexHeader
	statusCh chan *HexStatus
	quitCh   chan struct{}

	// Event handlers
	blockHandler  func(*hexcore.HexBlock) error
	headerHandler func(*hexcore.HexHeader) error
}

// HexMeshConfig contains configuration for the hex mesh protocol
type HexMeshConfig struct {
	NetworkID         uint64
	MaxPeers          int
	DialTimeout       time.Duration
	HandshakeTimeout  time.Duration
	PingInterval      time.Duration
	EnableNeighborOpt bool // Enable neighbor optimization
}

// DefaultHexMeshConfig returns default configuration
func DefaultHexMeshConfig() *HexMeshConfig {
	return &HexMeshConfig{
		NetworkID:         1337,
		MaxPeers:          50,
		DialTimeout:       30 * time.Second,
		HandshakeTimeout:  10 * time.Second,
		PingInterval:      15 * time.Second,
		EnableNeighborOpt: true,
	}
}

// HexPeer represents a connected peer in the hexagonal mesh
type HexPeer struct {
	id         enode.ID
	conn       *p2p.Peer
	rw         p2p.MsgReadWriter
	position   hexcore.HexCoordinate
	head       common.Hash
	difficulty uint64

	// Neighbor relationship
	isNeighbor bool
	distance   int64
	lastSeen   time.Time

	// Request tracking
	requests map[uint64]*PendingRequest
	reqMu    sync.RWMutex
	reqID    uint64
}

// PendingRequest tracks outgoing requests
type PendingRequest struct {
	ID        uint64
	Type      uint8
	Data      interface{}
	Timestamp time.Time
	Response  chan interface{}
}

// HexStatus represents the status message for handshake
type HexStatus struct {
	ProtocolVersion uint32                `json:"protocolVersion"`
	NetworkID       uint64                `json:"networkId"`
	Head            common.Hash           `json:"head"`
	Genesis         common.Hash           `json:"genesis"`
	Position        hexcore.HexCoordinate `json:"position"`
}

// NewHexMeshProtocol creates a new hex mesh protocol instance
func NewHexMeshProtocol(config *HexMeshConfig) *HexMeshProtocol {
	if config == nil {
		config = DefaultHexMeshConfig()
	}

	return &HexMeshProtocol{
		config:        config,
		peers:         make(map[enode.ID]*HexPeer),
		networkID:     config.NetworkID,
		localPosition: hexcore.NewHexCoordinate(0, 0), // Default position
		blockCh:       make(chan *hexcore.HexBlock, 100),
		headerCh:      make(chan *hexcore.HexHeader, 100),
		statusCh:      make(chan *HexStatus, 10),
		quitCh:        make(chan struct{}),
	}
}

// Start starts the hex mesh protocol
func (hmp *HexMeshProtocol) Start() error {
	log.Info("Starting Hexagonal Mesh Protocol", "version", HexMeshProtocolVersion)

	// Start background goroutines
	go hmp.heartbeatLoop()
	go hmp.messageHandler()

	return nil
}

// Stop stops the hex mesh protocol
func (hmp *HexMeshProtocol) Stop() {
	close(hmp.quitCh)

	// Disconnect all peers
	hmp.peersMu.Lock()
	for _, peer := range hmp.peers {
		peer.conn.Disconnect(p2p.DiscQuitting)
	}
	hmp.peersMu.Unlock()
}

// AddPeer adds a new peer to the mesh
func (hmp *HexMeshProtocol) AddPeer(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
	hexPeer := &HexPeer{
		id:       peer.ID(),
		conn:     peer,
		rw:       rw,
		requests: make(map[uint64]*PendingRequest),
		lastSeen: time.Now(),
	}

	// Perform handshake
	if err := hmp.handshake(hexPeer); err != nil {
		return fmt.Errorf("handshake failed: %v", err)
	}

	// Add to peers map
	hmp.peersMu.Lock()
	hmp.peers[peer.ID()] = hexPeer
	hmp.peersMu.Unlock()

	log.Info("Added hex mesh peer", "id", peer.ID().String()[:8], "position", hexPeer.position)

	// Start peer handler
	go hmp.handlePeer(hexPeer)

	return nil
}

// RemovePeer removes a peer from the mesh
func (hmp *HexMeshProtocol) RemovePeer(peerID enode.ID) {
	hmp.peersMu.Lock()
	delete(hmp.peers, peerID)
	hmp.peersMu.Unlock()

	log.Info("Removed hex mesh peer", "id", peerID.String()[:8])
}

// handshake performs the initial handshake with a peer
func (hmp *HexMeshProtocol) handshake(peer *HexPeer) error {
	// Send our status
	status := &HexStatus{
		ProtocolVersion: HexMeshProtocolVersion,
		NetworkID:       hmp.networkID,
		Head:            hmp.currentHead,
		Genesis:         common.Hash{}, // TODO: Get actual genesis hash
		Position:        hmp.localPosition,
	}

	if err := p2p.Send(peer.rw, HexStatusMsg, status); err != nil {
		return err
	}

	// Receive peer status
	msg, err := peer.rw.ReadMsg()
	if err != nil {
		return err
	}
	defer msg.Discard()

	if msg.Code != HexStatusMsg {
		return fmt.Errorf("expected status message, got %d", msg.Code)
	}

	var peerStatus HexStatus
	if err := msg.Decode(&peerStatus); err != nil {
		return err
	}

	// Validate peer status
	if peerStatus.NetworkID != hmp.networkID {
		return fmt.Errorf("network ID mismatch: got %d, want %d", peerStatus.NetworkID, hmp.networkID)
	}

	// Update peer information
	peer.position = peerStatus.Position
	peer.head = peerStatus.Head
	peer.distance = hmp.localPosition.Distance(peerStatus.Position)
	peer.isNeighbor = peer.distance == 1

	return nil
}

// handlePeer handles messages from a specific peer
func (hmp *HexMeshProtocol) handlePeer(peer *HexPeer) {
	defer func() {
		hmp.RemovePeer(peer.id)
		peer.conn.Disconnect(p2p.DiscSubprotocolError)
	}()

	for {
		msg, err := peer.rw.ReadMsg()
		if err != nil {
			log.Debug("Peer message read error", "peer", peer.id.String()[:8], "err", err)
			return
		}

		if err := hmp.handleMessage(peer, msg); err != nil {
			log.Debug("Failed to handle peer message", "peer", peer.id.String()[:8], "err", err)
			msg.Discard()
			return
		}
	}
}

// handleMessage handles a specific message from a peer
func (hmp *HexMeshProtocol) handleMessage(peer *HexPeer, msg p2p.Msg) error {
	defer msg.Discard()

	switch msg.Code {
	case HexBlockMsg:
		return hmp.handleHexBlock(peer, msg)
	case HexHeaderMsg:
		return hmp.handleHexHeader(peer, msg)
	case HexBlockRequestMsg:
		return hmp.handleBlockRequest(peer, msg)
	case HexHeaderRequestMsg:
		return hmp.handleHeaderRequest(peer, msg)
	case HexProofMsg:
		return hmp.handleHexProof(peer, msg)
	case HexNeighborMsg:
		return hmp.handleNeighborUpdate(peer, msg)
	case HexMeshStateMsg:
		return hmp.handleMeshState(peer, msg)
	default:
		return fmt.Errorf("unknown message code: %d", msg.Code)
	}
}

// handleHexBlock handles incoming hex blocks
func (hmp *HexMeshProtocol) handleHexBlock(peer *HexPeer, msg p2p.Msg) error {
	var block hexcore.HexBlock
	if err := msg.Decode(&block); err != nil {
		return err
	}

	peer.lastSeen = time.Now()

	// Send to block channel for processing
	select {
	case hmp.blockCh <- &block:
	default:
		log.Warn("Block channel full, dropping block", "hash", block.Hash().Hex()[:8])
	}

	// Call block handler if set
	if hmp.blockHandler != nil {
		return hmp.blockHandler(&block)
	}

	return nil
}

// handleHexHeader handles incoming hex headers
func (hmp *HexMeshProtocol) handleHexHeader(peer *HexPeer, msg p2p.Msg) error {
	var header hexcore.HexHeader
	if err := msg.Decode(&header); err != nil {
		return err
	}

	peer.lastSeen = time.Now()

	// Send to header channel for processing
	select {
	case hmp.headerCh <- &header:
	default:
		log.Warn("Header channel full, dropping header", "hash", header.Hash().Hex()[:8])
	}

	// Call header handler if set
	if hmp.headerHandler != nil {
		return hmp.headerHandler(&header)
	}

	return nil
}

// handleBlockRequest handles requests for specific blocks
func (hmp *HexMeshProtocol) handleBlockRequest(peer *HexPeer, msg p2p.Msg) error {
	var request struct {
		RequestID uint64      `json:"requestId"`
		Hash      common.Hash `json:"hash"`
	}

	if err := msg.Decode(&request); err != nil {
		return err
	}

	// TODO: Look up block and send response
	log.Debug("Received block request", "peer", peer.id.String()[:8], "hash", request.Hash.Hex()[:8])

	return nil
}

// handleHeaderRequest handles requests for specific headers
func (hmp *HexMeshProtocol) handleHeaderRequest(peer *HexPeer, msg p2p.Msg) error {
	var request struct {
		RequestID uint64      `json:"requestId"`
		Hash      common.Hash `json:"hash"`
	}

	if err := msg.Decode(&request); err != nil {
		return err
	}

	// TODO: Look up header and send response
	log.Debug("Received header request", "peer", peer.id.String()[:8], "hash", request.Hash.Hex()[:8])

	return nil
}

// handleHexProof handles hexagonal consensus proofs
func (hmp *HexMeshProtocol) handleHexProof(peer *HexPeer, msg p2p.Msg) error {
	var proof hexcore.HexaProof
	if err := msg.Decode(&proof); err != nil {
		return err
	}

	log.Debug("Received hex proof", "peer", peer.id.String()[:8], "hash", proof.Hash().Hex()[:8])

	// TODO: Validate and process proof
	return nil
}

// handleNeighborUpdate handles neighbor position updates
func (hmp *HexMeshProtocol) handleNeighborUpdate(peer *HexPeer, msg p2p.Msg) error {
	var update struct {
		Position hexcore.HexCoordinate `json:"position"`
		Head     common.Hash           `json:"head"`
	}

	if err := msg.Decode(&update); err != nil {
		return err
	}

	// Update peer information
	peer.position = update.Position
	peer.head = update.Head
	peer.distance = hmp.localPosition.Distance(update.Position)
	peer.isNeighbor = peer.distance == 1
	peer.lastSeen = time.Now()

	log.Debug("Updated peer position", "peer", peer.id.String()[:8], "position", update.Position)

	return nil
}

// handleMeshState handles mesh state synchronization
func (hmp *HexMeshProtocol) handleMeshState(peer *HexPeer, msg p2p.Msg) error {
	var state struct {
		KnownBlocks  []common.Hash `json:"knownBlocks"`
		KnownHeaders []common.Hash `json:"knownHeaders"`
	}

	if err := msg.Decode(&state); err != nil {
		return err
	}

	log.Debug("Received mesh state", "peer", peer.id.String()[:8], "blocks", len(state.KnownBlocks))

	// TODO: Process mesh state synchronization
	return nil
}

// BroadcastHexBlock broadcasts a hex block to relevant peers
func (hmp *HexMeshProtocol) BroadcastHexBlock(block *hexcore.HexBlock) {
	hmp.peersMu.RLock()
	defer hmp.peersMu.RUnlock()

	for _, peer := range hmp.peers {
		// Send to neighbors and close peers
		if peer.isNeighbor || peer.distance <= 3 {
			if err := p2p.Send(peer.rw, HexBlockMsg, block); err != nil {
				log.Debug("Failed to send block to peer", "peer", peer.id.String()[:8], "err", err)
			}
		}
	}
}

// BroadcastHexHeader broadcasts a hex header to relevant peers
func (hmp *HexMeshProtocol) BroadcastHexHeader(header *hexcore.HexHeader) {
	hmp.peersMu.RLock()
	defer hmp.peersMu.RUnlock()

	for _, peer := range hmp.peers {
		if err := p2p.Send(peer.rw, HexHeaderMsg, header); err != nil {
			log.Debug("Failed to send header to peer", "peer", peer.id.String()[:8], "err", err)
		}
	}
}

// SetLocalPosition sets the local node's position in the hex mesh
func (hmp *HexMeshProtocol) SetLocalPosition(pos hexcore.HexCoordinate) {
	hmp.localPosition = pos

	// Update neighbor relationships
	hmp.peersMu.Lock()
	for _, peer := range hmp.peers {
		peer.distance = hmp.localPosition.Distance(peer.position)
		peer.isNeighbor = peer.distance == 1
	}
	hmp.peersMu.Unlock()
}

// GetNeighborPeers returns peers that are direct neighbors
func (hmp *HexMeshProtocol) GetNeighborPeers() []*HexPeer {
	hmp.peersMu.RLock()
	defer hmp.peersMu.RUnlock()

	var neighbors []*HexPeer
	for _, peer := range hmp.peers {
		if peer.isNeighbor {
			neighbors = append(neighbors, peer)
		}
	}

	return neighbors
}

// heartbeatLoop sends periodic heartbeats and cleanups
func (hmp *HexMeshProtocol) heartbeatLoop() {
	ticker := time.NewTicker(time.Duration(HeartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hmp.sendHeartbeats()
			hmp.cleanupStaleRequests()
		case <-hmp.quitCh:
			return
		}
	}
}

// sendHeartbeats sends position updates to neighbors
func (hmp *HexMeshProtocol) sendHeartbeats() {
	update := struct {
		Position hexcore.HexCoordinate `json:"position"`
		Head     common.Hash           `json:"head"`
	}{
		Position: hmp.localPosition,
		Head:     hmp.currentHead,
	}

	hmp.peersMu.RLock()
	defer hmp.peersMu.RUnlock()

	for _, peer := range hmp.peers {
		if err := p2p.Send(peer.rw, HexNeighborMsg, update); err != nil {
			log.Debug("Failed to send heartbeat", "peer", peer.id.String()[:8], "err", err)
		}
	}
}

// cleanupStaleRequests removes old pending requests
func (hmp *HexMeshProtocol) cleanupStaleRequests() {
	now := time.Now()
	timeout := time.Duration(RequestTimeout) * time.Second

	hmp.peersMu.RLock()
	defer hmp.peersMu.RUnlock()

	for _, peer := range hmp.peers {
		peer.reqMu.Lock()
		for id, req := range peer.requests {
			if now.Sub(req.Timestamp) > timeout {
				close(req.Response)
				delete(peer.requests, id)
			}
		}
		peer.reqMu.Unlock()
	}
}

// messageHandler processes incoming messages from channels
func (hmp *HexMeshProtocol) messageHandler() {
	for {
		select {
		case block := <-hmp.blockCh:
			log.Debug("Processing hex block", "hash", block.Hash().Hex()[:8])
			// TODO: Process block
		case header := <-hmp.headerCh:
			log.Debug("Processing hex header", "hash", header.Hash().Hex()[:8])
			// TODO: Process header
		case status := <-hmp.statusCh:
			log.Debug("Processing status update", "networkID", status.NetworkID)
			// TODO: Process status
		case <-hmp.quitCh:
			return
		}
	}
}

// SetBlockHandler sets the handler for incoming blocks
func (hmp *HexMeshProtocol) SetBlockHandler(handler func(*hexcore.HexBlock) error) {
	hmp.blockHandler = handler
}

// SetHeaderHandler sets the handler for incoming headers
func (hmp *HexMeshProtocol) SetHeaderHandler(handler func(*hexcore.HexHeader) error) {
	hmp.headerHandler = handler
}

// GetProtocolSpec returns the P2P protocol specification
func (hmp *HexMeshProtocol) GetProtocolSpec() p2p.Protocol {
	return p2p.Protocol{
		Name:    HexMeshProtocolName,
		Version: HexMeshProtocolVersion,
		Length:  HexMeshProtocolLength,
		Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
			return hmp.AddPeer(peer, rw)
		},
	}
}
