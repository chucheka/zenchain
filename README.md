# Zenchain

> A blockchain built from scratch in Go — no libraries, no magic, just fundamentals.

Zenchain is a hands-on demo project that walks through the core mechanics of a blockchain ledger: accounts, transactions, state, and persistence. It is intentionally minimal so every concept is visible and understandable without wading through production-level abstractions.

---

## Suggested Name: **ChainForge**

`zenchain` works, but **ChainForge** better captures the spirit of the project — you are *forging* a blockchain from raw materials, hammering out each concept by hand. It signals to readers that this is a place where blockchains are built, not just used.

---

## What You Will Learn

- How a blockchain ledger tracks account **balances**
- How **transactions** are validated and applied to state
- How an **append-only transaction log** (`tx.db`) serves as the source of truth
- How **genesis state** seeds the initial chain
- How state is **reconstructed from scratch** by replaying the full transaction history
- How to wrap blockchain logic in a **CLI** using Cobra

---

## How It Works

```
genesis.json  ──┐
                ├──▶  NewStateFromDisk()  ──▶  State (in-memory balances)
tx.db (replay) ─┘                                      │
                                                        ▼
                                              Add(tx) + Persist()
                                                        │
                                                        ▼
                                               tx.db (appended)
```

1. On startup, genesis balances are loaded from `genesis.json`
2. Every transaction ever recorded in `tx.db` is replayed in order to rebuild current state
3. New transactions are validated against live balances, then appended to `tx.db`
4. `state.json` is a human-readable snapshot — it is derived, not authoritative

---

## Getting Started

**Prerequisites:** Go 1.22+

```bash
# Clone and enter the project
git clone https://github.com/chucheka/zenchain
cd zenchain

# Install dependencies
go mod download

# Build the CLI
go build -o zbb ./cmd/zenchain
```

---

## CLI Usage

```bash
# List all account balances
./zbb balances list

# Send tokens between accounts
./zbb tx add --from smith --to ola --value 500

# Issue a reward (mints tokens to an account, no sender deduction)
./zbb tx add --from smith --to smith --value 700 --data reward

# Show version
./zbb version
```

---

## Project Structure

```
zenchain/
├── cmd/zenchain/        # CLI entry point and Cobra commands
│   ├── main.go          # Root command setup
│   ├── balances.go      # `balances list` command
│   ├── tx.go            # `tx add` command
│   └── version.go       # `version` command
└── database/
    ├── tx.go            # Tx struct and Account type
    ├── genesis.go       # Genesis loader
    ├── state.go         # State management (apply, persist, replay)
    ├── genesis.json     # Initial account balances
    ├── tx.db            # Append-only transaction log (NDJSON)
    └── state.json       # Derived balance snapshot
```

---

## Key Concepts Illustrated

| Concept | Where to look |
|---|---|
| Transaction structure | `database/tx.go` |
| Genesis / initial state | `database/genesis.go`, `genesis.json` |
| State replay from log | `database/state.go` → `NewStateFromDisk()` |
| Balance validation | `database/state.go` → `apply()` |
| Reward transactions | `database/tx.go` → `IsReward()` |
| Append-only persistence | `database/state.go` → `Persist()` |
| CLI wiring | `cmd/zenchain/` |

---

## Concepts Not (Yet) Covered

This is a ledger, not a full blockchain. The following are natural next steps to extend the project:

- **Blocks** — grouping transactions into blocks with a block header
- **Hashing** — linking blocks via cryptographic hashes to form the chain
- **Proof of Work** — making block creation computationally expensive
- **Digital signatures** — proving transaction authenticity with public/private keys
- **Peer-to-peer networking** — syncing state across multiple nodes

---

## License

MIT
