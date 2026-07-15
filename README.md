# Golang Blockchain

This project is a simple blockchain implementation in Go. It stores runtime blockchain and wallet data under the tmp folder.

## Data rule
All files generated while running the blockchain nodes are considered temporary data and should not be committed to git. The tmp folder is therefore ignored by git so local chain state, wallet files, and Badger database files stay out of version control.

## Basic workflow
1. Create a wallet for a node.
2. Create the blockchain with a genesis reward to that wallet.
3. Start the node(s) with a NODE_ID value.
4. Send transactions and mine blocks.
5. Check balances or inspect the chain.

## How to run
Build the project:

```bash
go build ./...
```

Create a wallet:

```bash
NODE_ID=3000 go run main.go createwallet
```

Create the blockchain:

```bash
NODE_ID=3000 go run main.go createblockchain -address <wallet-address>
```

Start a node:

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

If you want to rebuild the UTXO set after changing local data:

```bash
NODE_ID=3000 go run main.go reindexutxo
```

