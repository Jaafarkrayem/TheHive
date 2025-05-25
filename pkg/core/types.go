// Package core contains the core types for Hexagonal Chain
package core

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

// HexCoordinate represents a position in hexagonal grid using cube coordinates
type HexCoordinate struct {
	Q int64 `json:"q"` // Column coordinate
	R int64 `json:"r"` // Row coordinate
	S int64 `json:"s"` // Calculated: S = -Q - R (cube coordinates)
}

// NewHexCoordinate creates a new hexagonal coordinate
func NewHexCoordinate(q, r int64) HexCoordinate {
	return HexCoordinate{
		Q: q,
		R: r,
		S: -q - r,
	}
}

// Distance calculates the distance between two hex coordinates
func (h HexCoordinate) Distance(other HexCoordinate) int64 {
	return (abs(h.Q-other.Q) + abs(h.R-other.R) + abs(h.S-other.S)) / 2
}

// Neighbors returns the 6 neighboring coordinates
func (h HexCoordinate) Neighbors() [6]HexCoordinate {
	directions := [6][2]int64{
		{1, 0},  // East
		{1, -1}, // NorthEast
		{0, -1}, // NorthWest
		{-1, 0}, // West
		{-1, 1}, // SouthWest
		{0, 1},  // SouthEast
	}

	var neighbors [6]HexCoordinate
	for i, dir := range directions {
		neighbors[i] = NewHexCoordinate(h.Q+dir[0], h.R+dir[1])
	}
	return neighbors
}

// HexDirection represents the 6 directions in a hexagonal grid
type HexDirection uint8

const (
	HexEast HexDirection = iota
	HexNorthEast
	HexNorthWest
	HexWest
	HexSouthWest
	HexSouthEast
)

// String returns the string representation of a hex direction
func (d HexDirection) String() string {
	directions := []string{"East", "NorthEast", "NorthWest", "West", "SouthWest", "SouthEast"}
	if int(d) < len(directions) {
		return directions[d]
	}
	return "Unknown"
}

// HexaProof contains consensus data for hexagonal validation
type HexaProof struct {
	NeighborSignatures [6][]byte        `json:"neighborSignatures"` // Signatures from neighbors
	StateProof         []byte           `json:"stateProof"`         // Proof of state consistency
	MeshProof          []byte           `json:"meshProof"`          // Proof of mesh integrity
	Timestamp          uint64           `json:"timestamp"`          // Consensus timestamp
	ValidatorSet       []common.Address `json:"validatorSet"`       // Active validators
	ProofHash          common.Hash      `json:"proofHash"`          // Hash of the proof
}

// Hash calculates the hash of the HexaProof
func (hp *HexaProof) Hash() common.Hash {
	if hp.ProofHash != (common.Hash{}) {
		return hp.ProofHash
	}

	// Create hash from all proof components
	hasher := sha256.New()
	for _, sig := range hp.NeighborSignatures {
		hasher.Write(sig)
	}
	hasher.Write(hp.StateProof)
	hasher.Write(hp.MeshProof)

	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, hp.Timestamp)
	hasher.Write(timestampBytes)

	for _, addr := range hp.ValidatorSet {
		hasher.Write(addr.Bytes())
	}

	copy(hp.ProofHash[:], hasher.Sum(nil))
	return hp.ProofHash
}

// HexHeader represents a hexagonal block header
type HexHeader struct {
	// Hexagonal-specific fields
	ParentHashes  [6]common.Hash `json:"parentHashes"`  // Up to 6 parent references
	NeighborCount uint8          `json:"neighborCount"` // Actual number of neighbors (0-6)
	HexPosition   HexCoordinate  `json:"hexPosition"`   // Position in hex grid
	MeshRoot      common.Hash    `json:"meshRoot"`      // State root across mesh
	HexProof      HexaProof      `json:"hexProof"`      // Consensus proof for neighbors

	// Standard Ethereum fields (inherited)
	Coinbase    common.Address   `json:"miner"`
	Root        common.Hash      `json:"stateRoot"`
	TxHash      common.Hash      `json:"transactionsRoot"`
	ReceiptHash common.Hash      `json:"receiptsRoot"`
	Bloom       types.Bloom      `json:"logsBloom"`
	Difficulty  *big.Int         `json:"difficulty"`
	Number      *big.Int         `json:"number"`
	GasLimit    uint64           `json:"gasLimit"`
	GasUsed     uint64           `json:"gasUsed"`
	Time        uint64           `json:"timestamp"`
	Extra       []byte           `json:"extraData"`
	MixDigest   common.Hash      `json:"mixHash"`
	Nonce       types.BlockNonce `json:"nonce"`

	// EIP fields
	BaseFee         *big.Int     `json:"baseFeePerGas,omitempty"`
	WithdrawalsHash *common.Hash `json:"withdrawalsRoot,omitempty"`
	BlobGasUsed     *uint64      `json:"blobGasUsed,omitempty"`
	ExcessBlobGas   *uint64      `json:"excessBlobGas,omitempty"`
}

// Hash calculates the hash of the hexagonal header
func (h *HexHeader) Hash() common.Hash {
	return rlpHash(h)
}

// ToEthHeader converts HexHeader to standard Ethereum Header for compatibility
func (h *HexHeader) ToEthHeader() *types.Header {
	// Use first non-zero parent as the primary parent
	var parentHash common.Hash
	for _, hash := range h.ParentHashes {
		if hash != (common.Hash{}) {
			parentHash = hash
			break
		}
	}

	return &types.Header{
		ParentHash:      parentHash,
		UncleHash:       types.EmptyUncleHash, // No uncles in hex chain
		Coinbase:        h.Coinbase,
		Root:            h.Root,
		TxHash:          h.TxHash,
		ReceiptHash:     h.ReceiptHash,
		Bloom:           h.Bloom,
		Difficulty:      h.Difficulty,
		Number:          h.Number,
		GasLimit:        h.GasLimit,
		GasUsed:         h.GasUsed,
		Time:            h.Time,
		Extra:           h.Extra,
		MixDigest:       h.MixDigest,
		Nonce:           h.Nonce,
		BaseFee:         h.BaseFee,
		WithdrawalsHash: h.WithdrawalsHash,
		BlobGasUsed:     h.BlobGasUsed,
		ExcessBlobGas:   h.ExcessBlobGas,
	}
}

// HexBlock represents a complete hexagonal block
type HexBlock struct {
	header       *HexHeader
	transactions []*types.Transaction
	withdrawals  []*types.Withdrawal

	// Hexagonal-specific data
	neighborProofs [6][]byte // Proofs from neighboring blocks
	meshWitness    []byte    // Witness data for mesh validation

	// Caches
	hash common.Hash
	size uint64

	// Tracking fields
	ReceivedAt   time.Time
	ReceivedFrom interface{}
}

// NewHexBlock creates a new hexagonal block
func NewHexBlock(header *HexHeader, txs []*types.Transaction, withdrawals []*types.Withdrawal) *HexBlock {
	return &HexBlock{
		header:       header,
		transactions: txs,
		withdrawals:  withdrawals,
	}
}

// Header returns the block header
func (b *HexBlock) Header() *HexHeader {
	return b.header
}

// Hash returns the block hash
func (b *HexBlock) Hash() common.Hash {
	if b.hash == (common.Hash{}) {
		b.hash = b.header.Hash()
	}
	return b.hash
}

// Number returns the block number
func (b *HexBlock) Number() *big.Int {
	return new(big.Int).Set(b.header.Number)
}

// Transactions returns the block transactions
func (b *HexBlock) Transactions() []*types.Transaction {
	return b.transactions
}

// Withdrawals returns the block withdrawals
func (b *HexBlock) Withdrawals() []*types.Withdrawal {
	return b.withdrawals
}

// ParentHashes returns all parent hashes
func (b *HexBlock) ParentHashes() [6]common.Hash {
	return b.header.ParentHashes
}

// HexPosition returns the hexagonal position
func (b *HexBlock) HexPosition() HexCoordinate {
	return b.header.HexPosition
}

// NeighborCount returns the number of neighbors
func (b *HexBlock) NeighborCount() uint8 {
	return b.header.NeighborCount
}

// ToEthBlock converts HexBlock to standard Ethereum Block for compatibility
func (b *HexBlock) ToEthBlock() *types.Block {
	ethHeader := b.header.ToEthHeader()
	body := &types.Body{
		Transactions: b.transactions,
		Withdrawals:  b.withdrawals,
	}
	return types.NewBlock(ethHeader, body, nil, nil)
}

// HexGenesisBlock creates the genesis block for hexagonal chain
func HexGenesisBlock() *HexBlock {
	header := &HexHeader{
		ParentHashes:  [6]common.Hash{}, // No parents for genesis
		NeighborCount: 0,
		HexPosition:   NewHexCoordinate(0, 0), // Origin position
		MeshRoot:      common.Hash{},
		HexProof:      HexaProof{},
		Coinbase:      common.Address{},
		Root:          common.Hash{},
		TxHash:        types.EmptyTxsHash,
		ReceiptHash:   types.EmptyReceiptsHash,
		Bloom:         types.Bloom{},
		Difficulty:    big.NewInt(1),
		Number:        big.NewInt(0),
		GasLimit:      5000000,
		GasUsed:       0,
		Time:          uint64(time.Now().Unix()),
		Extra:         []byte("Hexagonal Chain Genesis"),
		MixDigest:     common.Hash{},
		Nonce:         types.BlockNonce{},
	}

	return NewHexBlock(header, nil, nil)
}

// Helper functions

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func rlpHash(x interface{}) (h common.Hash) {
	hash := crypto.Keccak256Hash(rlpEncode(x))
	return hash
}

func rlpEncode(x interface{}) []byte {
	bytes, _ := rlp.EncodeToBytes(x)
	return bytes
}
