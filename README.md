# Golang Blockchain

This project is a simple blockchain implementation in Go with P2P networking between nodes. Each node is identified by a `NODE_ID` environment variable, which also serves as its TCP port (e.g. `NODE_ID=3000` listens on `localhost:3000`). It stores runtime blockchain and wallet data under the tmp folder.

## Data rule

All files generated while running the blockchain nodes are considered temporary data and should not be committed to git. The tmp folder is therefore ignored by git so local chain state, wallet files, and Badger database files stay out of version control.

## Prerequisites

Create the tmp folder before running any command. The app does not create it automatically and will fail with `open ./tmp/wallets_3000.data: no such file or directory` if it is missing:

```bash
mkdir -p tmp
```

Build the project:

```bash
go build ./...
```

## Basic workflow

1. Create a wallet for a node.
2. Create the blockchain with a genesis reward to that wallet.
3. Start the node(s) with a NODE_ID value.
4. Send transactions and mine blocks.
5. Check balances or inspect the chain.

## Single node quickstart

Create a wallet:

```bash
NODE_ID=3000 go run main.go createwallet
```

Create the blockchain (genesis reward goes to your wallet):

```bash
NODE_ID=3000 go run main.go createblockchain -address <wallet-address>
```

Start the node:

```bash
NODE_ID=3000 go run main.go startnode
```

Send a transaction and mine it immediately:

```bash
NODE_ID=3000 go run main.go send -from <from-address> -to <to-address> -amount 5 -mine
```

Check a balance:

```bash
NODE_ID=3000 go run main.go getbalance -address <wallet-address>
```

Print the chain:

```bash
NODE_ID=3000 go run main.go printchain
```

Rebuild the UTXO set after changing local data:

```bash
NODE_ID=3000 go run main.go reindexutxo
```

## Multi-node network (3 terminals)

Node `3000` is the hardcoded seed node (`KnownNodes[0]`); every other node syncs from it, so it must be started first.

### 1. Create one wallet per node

```bash
NODE_ID=3000 go run main.go createwallet   # -> ADDRESS_3000
NODE_ID=3001 go run main.go createwallet   # -> ADDRESS_3001
NODE_ID=3002 go run main.go createwallet   # -> ADDRESS_3002
```

### 2. Create the genesis block on the seed node only

```bash
NODE_ID=3000 go run main.go createblockchain -address <ADDRESS_3000>
```

### 3. Distribute the chain data to the other nodes

A node refuses to start without a local database (`No existing blockchain found, create one!`), and new blocks are only synced via the network after startup. Copy the seed node's database while **all nodes are stopped** (copying a live Badger database produces a corrupt/incomplete copy):

```bash
cp -r tmp/blocks_3000 tmp/blocks_3001
cp -r tmp/blocks_3000 tmp/blocks_3002
```

Never run `createblockchain` on 3001/3002 — there is only one genesis block, distributed to all peers.

### 4. Start the nodes, one per terminal

Terminal 1 — seed node:

```bash
NODE_ID=3000 go run main.go startnode
```

Terminal 2 — miner node (start only after 3000 is up):

```bash
NODE_ID=3001 go run main.go startnode -miner <ADDRESS_3001>
```

Terminal 3 — regular node:

```bash
NODE_ID=3002 go run main.go startnode
```

## Sending transactions between nodes

### The BadgerDB lock

A Badger database can only be opened by one process at a time. While a node is running, any CLI command using the same `NODE_ID` fails with `Cannot acquire directory lock on "./tmp/blocks_XXXX"`. Therefore, use a dedicated "CLI node" (`3003`) that is never started with `startnode`:

```bash
cp -r tmp/blocks_3000 tmp/blocks_3003       # do this while all nodes are stopped
cp tmp/wallets_3000.data tmp/wallets_3003.data
```

Transactions are created and signed locally, then sent to the seed node (`localhost:3000`), which broadcasts them to the network:

```bash
NODE_ID=3003 go run main.go send -from <ADDRESS_3000> -to <ADDRESS_3002> -amount 5
```

### Mining rule: the mempool needs 2 transactions

The miner only mines when its mempool holds **at least 2 transactions** (`len(memoryPool) >= 2` in `network/network.go`). A single transaction will sit in the mempool until a second one arrives, then both are mined in one block. Send two transactions back-to-back to trigger mining.

> Note: the CLI node's local database (3003) is not synced by the network. After a block is mined, its view of the UTXO set is stale, so a follow-up transaction from the same sender may attempt a double-spend. Re-sync it by stopping the seed node and copying `tmp/blocks_3000` again.

### Checking results

`getbalance` and `printchain` also open the database, so run them against a node that is currently stopped (Ctrl+C the node first, then restart it afterwards):

```bash
# Stop node 3002 (Ctrl+C), then:
NODE_ID=3002 go run main.go getbalance -address <ADDRESS_3002>
NODE_ID=3002 go run main.go printchain

# Start it again:
NODE_ID=3002 go run main.go startnode
```

## Command reference

| Command | Description |
|---|---|
| `createwallet` | Creates a new wallet and prints its address |
| `listaddresses` | Lists all addresses in the node's wallet file |
| `createblockchain -address ADDR` | Creates the genesis block, reward goes to ADDR |
| `startnode [-miner ADDR]` | Starts a node; `-miner` enables mining rewards to ADDR |
| `send -from F -to T -amount N [-mine]` | Sends N coins; `-mine` mines immediately on this node instead of broadcasting |
| `getbalance -address ADDR` | Prints the balance of ADDR |
| `printchain` | Prints all blocks in the local chain |
| `reindexutxo` | Rebuilds the UTXO set from the chain |

## Resetting everything

Stop all nodes, then wipe the tmp folder and start over:

```bash
rm -rf tmp && mkdir -p tmp
```
