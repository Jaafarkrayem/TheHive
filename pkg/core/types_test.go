package core

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestHexCoordinate(t *testing.T) {
	// Test coordinate creation
	coord := NewHexCoordinate(1, 2)
	if coord.Q != 1 || coord.R != 2 || coord.S != -3 {
		t.Errorf("NewHexCoordinate failed: got Q=%d, R=%d, S=%d, want Q=1, R=2, S=-3",
			coord.Q, coord.R, coord.S)
	}

	// Test distance calculation
	origin := NewHexCoordinate(0, 0)
	distance := origin.Distance(coord)
	expected := int64(3)
	if distance != expected {
		t.Errorf("Distance calculation failed: got %d, want %d", distance, expected)
	}

	// Test neighbors
	neighbors := origin.Neighbors()
	expectedNeighbors := [6]HexCoordinate{
		{1, 0, -1}, // East
		{1, -1, 0}, // NorthEast
		{0, -1, 1}, // NorthWest
		{-1, 0, 1}, // West
		{-1, 1, 0}, // SouthWest
		{0, 1, -1}, // SouthEast
	}

	for i, neighbor := range neighbors {
		if neighbor != expectedNeighbors[i] {
			t.Errorf("Neighbor %d failed: got %+v, want %+v",
				i, neighbor, expectedNeighbors[i])
		}
	}
}

func TestHexDirection(t *testing.T) {
	directions := []struct {
		dir      HexDirection
		expected string
	}{
		{HexEast, "East"},
		{HexNorthEast, "NorthEast"},
		{HexNorthWest, "NorthWest"},
		{HexWest, "West"},
		{HexSouthWest, "SouthWest"},
		{HexSouthEast, "SouthEast"},
	}

	for _, d := range directions {
		if d.dir.String() != d.expected {
			t.Errorf("Direction string failed: got %s, want %s",
				d.dir.String(), d.expected)
		}
	}
}

func TestHexaProof(t *testing.T) {
	proof := &HexaProof{
		NeighborSignatures: [6][]byte{
			[]byte("sig1"), []byte("sig2"), []byte("sig3"),
			[]byte("sig4"), []byte("sig5"), []byte("sig6"),
		},
		StateProof:   []byte("stateproof"),
		MeshProof:    []byte("meshproof"),
		Timestamp:    uint64(time.Now().Unix()),
		ValidatorSet: []common.Address{common.HexToAddress("0x1234")},
	}

	// Test hash generation
	hash1 := proof.Hash()
	hash2 := proof.Hash()

	// Hash should be consistent
	if hash1 != hash2 {
		t.Error("HexaProof hash is not consistent")
	}

	// Hash should not be empty
	if hash1 == (common.Hash{}) {
		t.Error("HexaProof hash should not be empty")
	}
}

func TestHexHeader(t *testing.T) {
	// Create test header
	header := &HexHeader{
		ParentHashes: [6]common.Hash{
			common.HexToHash("0x1234"),
			common.HexToHash("0x5678"),
		},
		NeighborCount: 2,
		HexPosition:   NewHexCoordinate(1, 1),
		MeshRoot:      common.HexToHash("0xabcd"),
		Coinbase:      common.HexToAddress("0x1234"),
		Root:          common.HexToHash("0xef01"),
		TxHash:        types.EmptyTxsHash,
		ReceiptHash:   types.EmptyReceiptsHash,
		Bloom:         types.Bloom{},
		Difficulty:    big.NewInt(1000),
		Number:        big.NewInt(1),
		GasLimit:      5000000,
		GasUsed:       100000,
		Time:          uint64(time.Now().Unix()),
		Extra:         []byte("test header"),
	}

	// Test hash generation
	hash := header.Hash()
	if hash == (common.Hash{}) {
		t.Error("Header hash should not be empty")
	}

	// Test conversion to Ethereum header
	ethHeader := header.ToEthHeader()
	if ethHeader.ParentHash != common.HexToHash("0x1234") {
		t.Error("Parent hash conversion failed")
	}
	if ethHeader.Number.Cmp(big.NewInt(1)) != 0 {
		t.Error("Block number conversion failed")
	}
	if ethHeader.GasLimit != 5000000 {
		t.Error("Gas limit conversion failed")
	}
}

func TestHexBlock(t *testing.T) {
	// Create test header
	header := &HexHeader{
		ParentHashes:  [6]common.Hash{common.HexToHash("0x1234")},
		NeighborCount: 1,
		HexPosition:   NewHexCoordinate(0, 1),
		MeshRoot:      common.HexToHash("0xabcd"),
		Coinbase:      common.HexToAddress("0x1234"),
		Root:          common.HexToHash("0xef01"),
		TxHash:        types.EmptyTxsHash,
		ReceiptHash:   types.EmptyReceiptsHash,
		Difficulty:    big.NewInt(1000),
		Number:        big.NewInt(1),
		GasLimit:      5000000,
		GasUsed:       0,
		Time:          uint64(time.Now().Unix()),
		Extra:         []byte("test block"),
	}

	// Create test block
	block := NewHexBlock(header, nil, nil)

	// Test basic accessors
	if block.Header() != header {
		t.Error("Header accessor failed")
	}

	if block.Number().Cmp(big.NewInt(1)) != 0 {
		t.Error("Number accessor failed")
	}

	if len(block.Transactions()) != 0 {
		t.Error("Transactions should be empty")
	}

	if len(block.Withdrawals()) != 0 {
		t.Error("Withdrawals should be empty")
	}

	// Test hex-specific accessors
	parentHashes := block.ParentHashes()
	if parentHashes[0] != common.HexToHash("0x1234") {
		t.Error("ParentHashes accessor failed")
	}

	if block.NeighborCount() != 1 {
		t.Error("NeighborCount accessor failed")
	}

	position := block.HexPosition()
	expectedPos := NewHexCoordinate(0, 1)
	if position != expectedPos {
		t.Errorf("HexPosition accessor failed: got %+v, want %+v",
			position, expectedPos)
	}

	// Test conversion to Ethereum block
	ethBlock := block.ToEthBlock()
	if ethBlock.Number().Cmp(big.NewInt(1)) != 0 {
		t.Error("Ethereum block conversion failed")
	}

	// Test hash generation
	hash := block.Hash()
	if hash == (common.Hash{}) {
		t.Error("Block hash should not be empty")
	}
}

func TestHexGenesisBlock(t *testing.T) {
	genesis := HexGenesisBlock()

	// Test genesis properties
	if genesis.Number().Cmp(big.NewInt(0)) != 0 {
		t.Error("Genesis block number should be 0")
	}

	if genesis.NeighborCount() != 0 {
		t.Error("Genesis block should have no neighbors")
	}

	position := genesis.HexPosition()
	origin := NewHexCoordinate(0, 0)
	if position != origin {
		t.Errorf("Genesis position should be origin: got %+v, want %+v",
			position, origin)
	}

	// All parent hashes should be empty
	parentHashes := genesis.ParentHashes()
	for i, hash := range parentHashes {
		if hash != (common.Hash{}) {
			t.Errorf("Genesis parent hash %d should be empty", i)
		}
	}

	// Test hash generation
	hash := genesis.Hash()
	if hash == (common.Hash{}) {
		t.Error("Genesis block hash should not be empty")
	}
}

func TestHexBlockWithTransactions(t *testing.T) {
	// Create test transaction
	tx := types.NewTransaction(
		0,                             // nonce
		common.HexToAddress("0x1234"), // to
		big.NewInt(1000),              // value
		21000,                         // gas limit
		big.NewInt(1000000000),        // gas price
		nil,                           // data
	)

	header := &HexHeader{
		ParentHashes:  [6]common.Hash{common.HexToHash("0x1234")},
		NeighborCount: 1,
		HexPosition:   NewHexCoordinate(1, 0),
		Number:        big.NewInt(1),
		GasLimit:      5000000,
		GasUsed:       21000,
		Time:          uint64(time.Now().Unix()),
	}

	block := NewHexBlock(header, []*types.Transaction{tx}, nil)

	// Test transaction inclusion
	if len(block.Transactions()) != 1 {
		t.Error("Block should contain one transaction")
	}

	if block.Transactions()[0] != tx {
		t.Error("Transaction should match the original")
	}
}

// Benchmark tests
func BenchmarkHexCoordinateDistance(b *testing.B) {
	coord1 := NewHexCoordinate(0, 0)
	coord2 := NewHexCoordinate(100, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		coord1.Distance(coord2)
	}
}

func BenchmarkHexHeaderHash(b *testing.B) {
	header := &HexHeader{
		ParentHashes:  [6]common.Hash{common.HexToHash("0x1234")},
		NeighborCount: 1,
		HexPosition:   NewHexCoordinate(1, 1),
		Number:        big.NewInt(1),
		GasLimit:      5000000,
		Time:          uint64(time.Now().Unix()),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		header.Hash()
	}
}

func BenchmarkHexaProofHash(b *testing.B) {
	proof := &HexaProof{
		NeighborSignatures: [6][]byte{
			[]byte("sig1"), []byte("sig2"), []byte("sig3"),
			[]byte("sig4"), []byte("sig5"), []byte("sig6"),
		},
		StateProof:   []byte("stateproof"),
		MeshProof:    []byte("meshproof"),
		Timestamp:    uint64(time.Now().Unix()),
		ValidatorSet: []common.Address{common.HexToAddress("0x1234")},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proof.ProofHash = common.Hash{} // Reset to force recalculation
		proof.Hash()
	}
}
