package database

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// newTestState builds a State directly using a temp file, bypassing NewStateFromDisk.
// The caller should defer the returned cleanup function.
func newTestState(t *testing.T, balances map[Account]uint) (*State, func()) {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "block-*.db")
	if err != nil {
		t.Fatal(err)
	}
	s := &State{
		Balances:  balances,
		txMempool: []Tx{},
		dbFile:    f,
	}
	return s, func() { f.Close() }
}

// --- apply() ---

func TestApply_Reward(t *testing.T) {
	s, cleanup := newTestState(t, map[Account]uint{"smith": 0})
	defer cleanup()

	tx := NewTx("smith", "smith", 700, "reward")
	if err := s.apply(tx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Balances["smith"] != 700 {
		t.Errorf("expected balance 700, got %d", s.Balances["smith"])
	}
}

func TestApply_Reward_DoesNotDeductSender(t *testing.T) {
	// Reward should credit the recipient without touching the sender's balance.
	s, cleanup := newTestState(t, map[Account]uint{"smith": 100})
	defer cleanup()

	tx := NewTx("smith", "ola", 500, "reward")
	if err := s.apply(tx); err != nil {
		t.Fatal(err)
	}
	if s.Balances["smith"] != 100 {
		t.Errorf("sender should be unchanged; expected 100, got %d", s.Balances["smith"])
	}
	if s.Balances["ola"] != 500 {
		t.Errorf("expected ola=500, got %d", s.Balances["ola"])
	}
}

func TestApply_RegularTx(t *testing.T) {
	s, cleanup := newTestState(t, map[Account]uint{"smith": 1000, "ola": 0})
	defer cleanup()

	tx := NewTx("smith", "ola", 300, "")
	if err := s.apply(tx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Balances["smith"] != 700 {
		t.Errorf("expected smith=700, got %d", s.Balances["smith"])
	}
	if s.Balances["ola"] != 300 {
		t.Errorf("expected ola=300, got %d", s.Balances["ola"])
	}
}

func TestApply_InsufficientBalance(t *testing.T) {
	s, cleanup := newTestState(t, map[Account]uint{"smith": 10})
	defer cleanup()

	tx := NewTx("smith", "ola", 100, "")
	err := s.apply(tx)
	if err == nil {
		t.Error("expected error for insufficient balance, got nil")
	}
}

func TestApply_ExactBalance(t *testing.T) {
	s, cleanup := newTestState(t, map[Account]uint{"smith": 100, "ola": 0})
	defer cleanup()

	tx := NewTx("smith", "ola", 100, "")
	if err := s.apply(tx); err != nil {
		t.Fatalf("unexpected error spending exact balance: %v", err)
	}
	if s.Balances["smith"] != 0 {
		t.Errorf("expected smith=0, got %d", s.Balances["smith"])
	}
}

func TestApply_NewRecipientCreated(t *testing.T) {
	s, cleanup := newTestState(t, map[Account]uint{"smith": 500})
	defer cleanup()

	tx := NewTx("smith", "caesar", 200, "")
	if err := s.apply(tx); err != nil {
		t.Fatal(err)
	}
	if s.Balances["caesar"] != 200 {
		t.Errorf("expected caesar=200, got %d", s.Balances["caesar"])
	}
}

// --- AddTx() ---

func TestAddTx_AppliesAndQueues(t *testing.T) {
	s, cleanup := newTestState(t, map[Account]uint{"smith": 1000})
	defer cleanup()

	tx := NewTx("smith", "ola", 250, "")
	if err := s.AddTx(tx); err != nil {
		t.Fatal(err)
	}
	if s.Balances["smith"] != 750 {
		t.Errorf("expected smith=750, got %d", s.Balances["smith"])
	}
	if len(s.txMempool) != 1 {
		t.Errorf("expected 1 tx in mempool, got %d", len(s.txMempool))
	}
}

func TestAddTx_RejectsInvalidTx(t *testing.T) {
	s, cleanup := newTestState(t, map[Account]uint{"smith": 10})
	defer cleanup()

	tx := NewTx("smith", "ola", 9999, "")
	if err := s.AddTx(tx); err == nil {
		t.Error("expected error for overspend, got nil")
	}
	if len(s.txMempool) != 0 {
		t.Errorf("invalid tx should not reach mempool, got %d", len(s.txMempool))
	}
}

// --- AddBlock() ---

func TestAddBlock_AllTxsApplied(t *testing.T) {
	s, cleanup := newTestState(t, map[Account]uint{"smith": 1000})
	defer cleanup()

	block := NewBlock(Hash{}, 1729000000, []Tx{
		NewTx("smith", "smith", 100, "reward"),
		NewTx("smith", "ola", 200, ""),
	})

	if err := s.AddBlock(block); err != nil {
		t.Fatal(err)
	}
	if s.Balances["smith"] != 900 {
		t.Errorf("expected smith=900 (1000-200+reward credit=100), got %d", s.Balances["smith"])
	}
	if s.Balances["ola"] != 200 {
		t.Errorf("expected ola=200, got %d", s.Balances["ola"])
	}
}

func TestAddBlock_FailsOnBadTx(t *testing.T) {
	s, cleanup := newTestState(t, map[Account]uint{"smith": 50})
	defer cleanup()

	block := NewBlock(Hash{}, 1729000000, []Tx{
		NewTx("smith", "ola", 9999, ""),
	})

	if err := s.AddBlock(block); err == nil {
		t.Error("expected error when block contains an invalid tx")
	}
}

// --- Persist() ---

func TestPersist_WritesBlockAndClearsMempool(t *testing.T) {
	s, cleanup := newTestState(t, map[Account]uint{"smith": 1000})
	defer cleanup()

	_ = s.AddTx(NewTx("smith", "ola", 100, ""))

	hash, err := s.Persist()
	if err != nil {
		t.Fatal(err)
	}

	if hash == (Hash{}) {
		t.Error("persisted hash should not be zero")
	}
	if s.latestBlockHash != hash {
		t.Error("latestBlockHash not updated after persist")
	}
	if len(s.txMempool) != 0 {
		t.Errorf("mempool should be empty after persist, got %d", len(s.txMempool))
	}
}

func TestPersist_BlockChaining(t *testing.T) {
	s, cleanup := newTestState(t, map[Account]uint{"smith": 1000})
	defer cleanup()

	_ = s.AddTx(NewTx("smith", "ola", 10, ""))
	hash0, err := s.Persist()
	if err != nil {
		t.Fatal(err)
	}

	_ = s.AddTx(NewTx("smith", "ola", 10, ""))
	hash1, err := s.Persist()
	if err != nil {
		t.Fatal(err)
	}

	if hash0 == hash1 {
		t.Error("sequential blocks must have different hashes")
	}
}

func TestPersist_BlockWrittenToFile(t *testing.T) {
	s, cleanup := newTestState(t, map[Account]uint{"smith": 500})
	defer cleanup()

	_ = s.AddTx(NewTx("smith", "ola", 50, ""))
	if _, err := s.Persist(); err != nil {
		t.Fatal(err)
	}

	// Rewind and read the file to verify a block was written
	if _, err := s.dbFile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(s.dbFile)
	if !scanner.Scan() {
		t.Fatal("expected at least one line in db file, got none")
	}

	var blockFs BlockFS
	if err := json.Unmarshal(scanner.Bytes(), &blockFs); err != nil {
		t.Fatalf("could not unmarshal persisted block: %v", err)
	}
	if len(blockFs.Value.TXs) != 1 {
		t.Errorf("expected 1 tx in persisted block, got %d", len(blockFs.Value.TXs))
	}
}

// --- Integration: NewStateFromDisk ---

// setupStateDir creates a temp directory mimicking the project root layout:
//
//	<tmp>/
//	  database/
//	    genesis.json
//	    block.db
//
// It returns the root directory. Callers should os.Chdir into it before
// calling NewStateFromDisk (and restore CWD afterward).
func setupStateDir(t *testing.T, genesisJSON string, blockLines []string) string {
	t.Helper()
	root := t.TempDir()
	dbDir := filepath.Join(root, "database")
	if err := os.Mkdir(dbDir, 0700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dbDir, "genesis.json"), []byte(genesisJSON), 0600); err != nil {
		t.Fatal(err)
	}

	var blockContent []byte
	for _, line := range blockLines {
		blockContent = append(blockContent, []byte(line+"\n")...)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "block.db"), blockContent, 0600); err != nil {
		t.Fatal(err)
	}

	return root
}

func TestNewStateFromDisk_GenesisOnly(t *testing.T) {
	root := setupStateDir(t,
		`{"balances":{"smith":1000000}}`,
		nil,
	)

	orig, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	state, err := NewStateFromDisk()
	if err != nil {
		t.Fatal(err)
	}
	defer state.Close()

	if state.Balances["smith"] != 1000000 {
		t.Errorf("expected smith=1000000, got %d", state.Balances["smith"])
	}
	if state.LatestBlockHash() != (Hash{}) {
		t.Error("expected zero latest block hash when no blocks exist")
	}
}

func TestNewStateFromDisk_ReplaysBlocks(t *testing.T) {
	// Persist a block manually via the test state, capture the JSON line,
	// then verify NewStateFromDisk replays it correctly.
	tmp := t.TempDir()
	f, err := os.CreateTemp(tmp, "block-*.db")
	if err != nil {
		t.Fatal(err)
	}

	s := &State{
		Balances:  map[Account]uint{"smith": 1000000},
		txMempool: []Tx{},
		dbFile:    f,
	}
	_ = s.AddTx(NewTx("smith", "ola", 2000, ""))
	_ = s.AddTx(NewTx("smith", "smith", 100, "reward"))
	if _, err := s.Persist(); err != nil {
		t.Fatal(err)
	}

	// Read back what was written
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	f.Close()

	root := setupStateDir(t, `{"balances":{"smith":1000000}}`, lines)

	orig, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	state, err := NewStateFromDisk()
	if err != nil {
		t.Fatal(err)
	}
	defer state.Close()

	// smith: 1000000 - 2000 + 100 (reward) = 998100
	if state.Balances["smith"] != 998100 {
		t.Errorf("expected smith=998100, got %d", state.Balances["smith"])
	}
	if state.Balances["ola"] != 2000 {
		t.Errorf("expected ola=2000, got %d", state.Balances["ola"])
	}
	if state.LatestBlockHash() == (Hash{}) {
		t.Error("expected non-zero latest block hash after replaying blocks")
	}
}

func TestNewStateFromDisk_MissingGenesisFile(t *testing.T) {
	root := t.TempDir()
	// Create database dir but omit genesis.json
	if err := os.Mkdir(filepath.Join(root, "database"), 0700); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	_, err := NewStateFromDisk()
	if err == nil {
		t.Error("expected error when genesis.json is missing")
	}
}
