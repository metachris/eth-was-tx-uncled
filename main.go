package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var ethNodeUriPtr = flag.String("eth", os.Getenv("ETH_NODE_URI"), "URL for eth node (eth node, Infura, etc.)")
var txHashPtr = flag.String("tx", "", "tx hash")

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

	// TODO: if this block has no uncles, perhaps try a few blocks back?
	for _, uncle := range block.Uncles() {
		uncleBlock, err := client.BlockByHash(context.Background(), uncle.Hash())
		Perror(err)

		for _, uncleTx := range uncleBlock.Transactions() {
			if tx.Hash() == uncleTx.Hash() {
				fmt.Printf("Found tx in uncle %s\n", uncle.Hash().Hex())
				return
			}
		}
	}

	fmt.Printf("Tx %s not found in uncles\n", tx.Hash().Hex())
}
