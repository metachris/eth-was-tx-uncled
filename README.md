Find out if a tx was uncled (first mined in a block that has been uncled, in which case the tx was sent back to mempool and mined again later).

## Usage:

```bash
go run cmd/cli/main.go -eth ETH_NODE_URI -tx TX_HASH
```

You can also use the environment variable `ETH_NODE_URI` to set a default node URI.

## Installation:

```bash
# Install
go install github.com/metachris/eth-was-tx-uncled@master

# Use
eth-was-tx-uncled go run . -eth ETH_NODE_URI -tx TX_HASH
```
