package database

import (
	"encoding/hex"
	"testing"
)

func TestNewBlock(t *testing.T) {
	parent := Hash{}
	txs := []Tx{NewTx("smith", "ola", 100, "")}
	block := NewBlock(parent, 1729000000, txs)

	if block.Header.Parent != parent {
		t.Errorf("expected zero parent hash, got %v", block.Header.Parent)
	}
	if block.Header.Time != 1729000000 {
		t.Errorf("expected Time=1729000000, got %d", block.Header.Time)
	}
	if len(block.TXs) != 1 {
		t.Errorf("expected 1 tx, got %d", len(block.TXs))
	}
}

func TestBlockHash_Deterministic(t *testing.T) {
	block := NewBlock(Hash{}, 1729000000, []Tx{NewTx("smith", "ola", 50, "")})

	h1, err := block.Hash()
	if err != nil {
		t.Fatal(err)
	}
	h2, err := block.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Error("same block should produce the same hash each time")
	}
}

func TestBlockHash_DifferentBlocks(t *testing.T) {
	b1 := NewBlock(Hash{}, 1729000000, []Tx{NewTx("smith", "ola", 50, "")})
	b2 := NewBlock(Hash{}, 1729000001, []Tx{NewTx("smith", "ola", 50, "")})

	h1, _ := b1.Hash()
	h2, _ := b2.Hash()
	if h1 == h2 {
		t.Error("different blocks should produce different hashes")
	}
}

func TestBlockHash_NonZero(t *testing.T) {
	block := NewBlock(Hash{}, 1729000000, []Tx{NewTx("smith", "ola", 1, "")})
	h, err := block.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if h == (Hash{}) {
		t.Error("block hash should not be a zero hash")
	}
}

func TestHashMarshalUnmarshal(t *testing.T) {
	// Build a known non-zero hash
	var h Hash
	src, _ := hex.DecodeString("79d5f855bb8241d022a3e36a346c1ae92b85a18c08f0a12ea4e086017500c13d")
	copy(h[:], src)

	text, err := h.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	var decoded Hash
	if err := decoded.UnmarshalText(text); err != nil {
		t.Fatal(err)
	}

	if h != decoded {
		t.Errorf("round-trip failed: got %x, want %x", decoded, h)
	}
}

func TestHashMarshalText_ZeroHash(t *testing.T) {
	var h Hash
	text, err := h.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	expected := "0000000000000000000000000000000000000000000000000000000000000000"
	if string(text) != expected {
		t.Errorf("expected all-zeros hex, got %q", text)
	}
}
