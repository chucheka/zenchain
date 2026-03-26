package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGenesis_Valid(t *testing.T) {
	content := `{
		"genesis_time": "2024-10-15T00:00:00.000000000Z",
		"chain_id": "test-chain",
		"balances": {
			"smith": 1000000,
			"ola": 5000
		}
	}`

	path := filepath.Join(t.TempDir(), "genesis.json")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	gen, err := loadGenesis(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gen.Balances["smith"] != 1000000 {
		t.Errorf("expected smith=1000000, got %d", gen.Balances["smith"])
	}
	if gen.Balances["ola"] != 5000 {
		t.Errorf("expected ola=5000, got %d", gen.Balances["ola"])
	}
}

func TestLoadGenesis_FileNotFound(t *testing.T) {
	_, err := loadGenesis("/nonexistent/path/genesis.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadGenesis_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "genesis.json")
	if err := os.WriteFile(path, []byte(`{invalid json`), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := loadGenesis(path)
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

func TestLoadGenesis_EmptyBalances(t *testing.T) {
	content := `{"balances": {}}`
	path := filepath.Join(t.TempDir(), "genesis.json")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	gen, err := loadGenesis(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gen.Balances) != 0 {
		t.Errorf("expected 0 balances, got %d", len(gen.Balances))
	}
}
