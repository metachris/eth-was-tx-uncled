package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var ethNodeUriPtr = flag.String("eth", os.Getenv("ETH_NODE_URI"), "URL for eth node (eth node, Infura, etc.)")
var txHashPtr = flag.String("tx", "", "tx hash")
var debugPtr = flag.Bool("debug", false, "print additional information")

func Perror(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	if *ethNodeUriPtr == "" {
		panic("No eth node URI provided")
	}

	if *txHashPtr == "" {
		panic("No eth node URI provided")
	}

	client, err := ethclient.Dial(*ethNodeUriPtr)
	Perror(err)

	tx, _, err := client.TransactionByHash(context.Background(), common.HexToHash(*txHashPtr))
	if err == ethereum.NotFound {
		fmt.Println("Transaction not found")
		return
	}
	Perror(err)

	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	Perror(err)

	block, err := client.BlockByHash(context.Background(), receipt.BlockHash)
	Perror(err)

	// check uncles of included block, and if not found check the previous few blocks for uncles
	for blocksTried := 0; blocksTried < 6; blocksTried++ {
		found, blockHash := wasTxFoundInUncles(client, block, tx.Hash())
		if found {
			fmt.Printf("Found tx in uncle %s\n", blockHash.Hex())
			return
		}
		prevBlockNumber := big.NewInt(block.Number().Int64() - 1)
		block, err = client.BlockByNumber(context.Background(), prevBlockNumber)
		Perror(err)
	}

	fmt.Printf("Tx %s not found in uncles\n", tx.Hash().Hex())
}

func wasTxFoundInUncles(client *ethclient.Client, block *types.Block, txHash common.Hash) (found bool, blockHash common.Hash) {
	if *debugPtr {
		fmt.Println("checking block", block.Number().Int64())
	}
	for _, uncle := range block.Uncles() {
		uncleBlock, err := client.BlockByHash(context.Background(), uncle.Hash())
		Perror(err)
		if *debugPtr {
			fmt.Println("- checking uncle", uncleBlock.Hash().Hex())
		}

		for _, uncleTx := range uncleBlock.Transactions() {
			if txHash == uncleTx.Hash() {
				return true, uncle.Hash()
			}
		}
	}
	return false, blockHash
}
