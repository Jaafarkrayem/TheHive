// Package config provides configuration management for the Hexagonal Chain node
package config

import (
	"fmt"
	"time"
)

// Config represents the complete configuration for a Hexagonal Chain node
type Config struct {
	// Node configuration
	DataDir   string `json:"datadir"`
	NetworkID uint64 `json:"networkid"`
	NodeType  string `json:"nodetype"` // "full", "light", "validator"
	Validator bool   `json:"validator"`

	// HTTP API configuration
	HTTP HTTPConfig `json:"http"`

	// WebSocket API configuration
	WebSocket WSConfig `json:"websocket"`

	// P2P network configuration
	P2P P2PConfig `json:"p2p"`

	// Hexagonal Chain specific configuration
	HexChain HexChainConfig `json:"hexchain"`

	// Consensus configuration
	Consensus ConsensusConfig `json:"consensus"`

	// Mining configuration
	Mining MiningConfig `json:"mining"`
}

// HTTPConfig represents HTTP API configuration
type HTTPConfig struct {
	Enabled bool   `json:"enabled"`
	Addr    string `json:"addr"`
	Port    int    `json:"port"`
	CORS    string `json:"cors"`
	VHosts  string `json:"vhosts"`
}

// WSConfig represents WebSocket API configuration
type WSConfig struct {
	Enabled bool   `json:"enabled"`
	Addr    string `json:"addr"`
	Port    int    `json:"port"`
	Origins string `json:"origins"`
}

// P2PConfig represents P2P network configuration
type P2PConfig struct {
	Port           int      `json:"port"`
	MaxPeers       int      `json:"maxpeers"`
	BootstrapNodes []string `json:"bootstrapnodes"`
	Discovery      bool     `json:"discovery"`
	NAT            bool     `json:"nat"`
}

// HexChainConfig represents Hexagonal Chain specific configuration
type HexChainConfig struct {
	// Mesh topology settings
	MaxNeighbors     int  `json:"maxneighbors"`     // Maximum number of neighbors (1-6)
	MinNeighbors     int  `json:"minneighbors"`     // Minimum neighbors for finality
	MeshOptimization bool `json:"meshoptimization"` // Enable dynamic mesh optimization

	// Block production settings
	BlockTime          time.Duration `json:"blocktime"`       // Target block time
	ParentSelectionAlg string        `json:"parentselection"` // Algorithm for parent selection

	// State management
	StateSync    bool `json:"statesync"`    // Enable state synchronization
	StatePruning bool `json:"statepruning"` // Enable state pruning

	// Hexagonal coordinate settings
	InitialPosition  HexPosition `json:"initialposition"`  // Initial hex position
	PositionStrategy string      `json:"positionstrategy"` // Position selection strategy
}

// HexPosition represents a position in the hexagonal grid
type HexPosition struct {
	Q int64 `json:"q"`
	R int64 `json:"r"`
	S int64 `json:"s"`
}

// ConsensusConfig represents consensus mechanism configuration
type ConsensusConfig struct {
	Algorithm string `json:"algorithm"` // "hexaproof", "poa", "pos"

	// HexaProof specific settings
	FinalizationTime time.Duration `json:"finalizationtime"` // Time to wait for finalization
	SignatureTimeout time.Duration `json:"signaturetimeout"` // Timeout for neighbor signatures
	ConflictResolver string        `json:"conflictresolver"` // Algorithm for conflict resolution

	// Validator settings
	ValidatorTimeout time.Duration `json:"validatortimeout"` // Validator response timeout
	RequiredSigners  int           `json:"requiredsigners"`  // Required number of signers
}

// MiningConfig represents mining/block production configuration
type MiningConfig struct {
	Enabled  bool          `json:"enabled"`
	Threads  int           `json:"threads"`
	GasFloor uint64        `json:"gasfloor"`
	GasCeil  uint64        `json:"gasceil"`
	GasPrice uint64        `json:"gasprice"`
	Recommit time.Duration `json:"recommit"`

	// Hex-specific mining settings
	HexMining bool `json:"hexmining"` // Enable hexagonal mining
	MeshAware bool `json:"meshaware"` // Consider mesh topology in mining
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		DataDir:   "./data",
		NetworkID: 1337,
		NodeType:  "full",
		Validator: false,

		HTTP: HTTPConfig{
			Enabled: true,
			Addr:    "localhost",
			Port:    8545,
			CORS:    "*",
			VHosts:  "*",
		},

		WebSocket: WSConfig{
			Enabled: true,
			Addr:    "localhost",
			Port:    8546,
			Origins: "*",
		},

		P2P: P2PConfig{
			Port:           30303,
			MaxPeers:       50,
			BootstrapNodes: []string{},
			Discovery:      true,
			NAT:            true,
		},

		HexChain: HexChainConfig{
			MaxNeighbors:       6,
			MinNeighbors:       3,
			MeshOptimization:   true,
			BlockTime:          2 * time.Second,
			ParentSelectionAlg: "optimal",
			StateSync:          true,
			StatePruning:       true,
			InitialPosition:    HexPosition{Q: 0, R: 0, S: 0},
			PositionStrategy:   "auto",
		},

		Consensus: ConsensusConfig{
			Algorithm:        "hexaproof",
			FinalizationTime: 6 * time.Second,
			SignatureTimeout: 1 * time.Second,
			ConflictResolver: "weighted",
			ValidatorTimeout: 2 * time.Second,
			RequiredSigners:  3,
		},

		Mining: MiningConfig{
			Enabled:   false,
			Threads:   1,
			GasFloor:  8000000,
			GasCeil:   8000000,
			GasPrice:  1000000000,
			Recommit:  2 * time.Second,
			HexMining: true,
			MeshAware: true,
		},
	}
}

// ValidatorConfig returns a configuration optimized for validator nodes
func ValidatorConfig() *Config {
	cfg := DefaultConfig()
	cfg.NodeType = "validator"
	cfg.Validator = true
	cfg.Mining.Enabled = true
	cfg.P2P.MaxPeers = 100
	cfg.HexChain.MinNeighbors = 4 // Require more neighbors for validators
	return cfg
}

// LightConfig returns a configuration optimized for light clients
func LightConfig() *Config {
	cfg := DefaultConfig()
	cfg.NodeType = "light"
	cfg.P2P.MaxPeers = 10
	cfg.HexChain.MaxNeighbors = 2 // Fewer neighbors for light clients
	cfg.HexChain.MinNeighbors = 1
	cfg.HexChain.StateSync = false
	cfg.Mining.Enabled = false
	return cfg
}

// TestnetConfig returns a configuration for testnet
func TestnetConfig() *Config {
	cfg := DefaultConfig()
	cfg.NetworkID = 1337
	cfg.HexChain.BlockTime = 1 * time.Second // Faster blocks for testing
	cfg.Consensus.FinalizationTime = 3 * time.Second
	return cfg
}

// MainnetConfig returns a configuration for mainnet
func MainnetConfig() *Config {
	cfg := DefaultConfig()
	cfg.NetworkID = 1
	cfg.HexChain.BlockTime = 12 * time.Second
	cfg.Consensus.FinalizationTime = 12 * time.Second
	cfg.HexChain.MinNeighbors = 4 // Higher security for mainnet
	return cfg
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check hex chain constraints
	if c.HexChain.MaxNeighbors > 6 {
		return fmt.Errorf("maxneighbors cannot exceed 6")
	}
	if c.HexChain.MinNeighbors > c.HexChain.MaxNeighbors {
		return fmt.Errorf("minneighbors cannot exceed maxneighbors")
	}
	if c.HexChain.MinNeighbors < 1 {
		return fmt.Errorf("minneighbors must be at least 1")
	}

	// Check consensus settings
	if c.Consensus.RequiredSigners > c.HexChain.MaxNeighbors {
		return fmt.Errorf("requiredsigners cannot exceed maxneighbors")
	}

	// Check mining settings
	if c.Mining.Enabled && c.Mining.Threads < 1 {
		return fmt.Errorf("mining threads must be at least 1")
	}

	return nil
}

// SetBootstrapNodes sets the bootstrap nodes for P2P discovery
func (c *Config) SetBootstrapNodes(nodes []string) {
	c.P2P.BootstrapNodes = nodes
}

// IsValidator returns true if this node is configured as a validator
func (c *Config) IsValidator() bool {
	return c.Validator || c.NodeType == "validator"
}

// IsLightClient returns true if this node is configured as a light client
func (c *Config) IsLightClient() bool {
	return c.NodeType == "light"
}
