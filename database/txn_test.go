package database

import "testing"

func TestNewAccount(t *testing.T) {
	acc := NewAccount("smith")
	if acc != Account("smith") {
		t.Errorf("expected Account(\"smith\"), got %q", acc)
	}
}

func TestNewTx(t *testing.T) {
	tx := NewTx("smith", "ola", 500, "")
	if tx.From != "smith" {
		t.Errorf("expected From=smith, got %q", tx.From)
	}
	if tx.To != "ola" {
		t.Errorf("expected To=ola, got %q", tx.To)
	}
	if tx.Value != 500 {
		t.Errorf("expected Value=500, got %d", tx.Value)
	}
	if tx.Data != "" {
		t.Errorf("expected empty Data, got %q", tx.Data)
	}
}

func TestTxIsReward(t *testing.T) {
	reward := NewTx("smith", "smith", 100, "reward")
	if !reward.IsReward() {
		t.Error("expected IsReward()=true for data=\"reward\"")
	}
}

func TestTxIsNotReward(t *testing.T) {
	tx := NewTx("smith", "ola", 100, "")
	if tx.IsReward() {
		t.Error("expected IsReward()=false for empty data")
	}

	tx2 := NewTx("smith", "ola", 100, "transfer")
	if tx2.IsReward() {
		t.Error("expected IsReward()=false for data=\"transfer\"")
	}
}
