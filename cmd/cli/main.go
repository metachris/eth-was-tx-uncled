package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/metachris/eth-was-tx-uncled/txinfo"
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

	status, uncleBlock, err := txinfo.WasTxUncled(client, common.HexToHash(*txHashPtr))
	Perror(err)

	if status == txinfo.StatusTxUnknown {
		fmt.Println("tx not found")
	} else if status == txinfo.StatusTxNotUncled {
		fmt.Println("tx not uncled")
	} else if status == txinfo.StatusTxWasUncled {
		fmt.Printf("tx was uncled in block %s %s\n", uncleBlock.Number(), uncleBlock.Hash().Hex())
	}
}
