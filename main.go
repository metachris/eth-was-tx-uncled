package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/metachris/eth-was-tx-uncled/txinfo"
)

var (
	ethNodeUriPtr = flag.String("eth", os.Getenv("ETH_NODE_URI"), "URL for eth node (eth node, Infura, etc.)")
	txHashPtr     = flag.String("tx", "", "tx hash")
)

func main() {
	flag.Parse()
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))

	if *ethNodeUriPtr == "" {
		log.Crit("No eth node URI provided")
	}

	if *txHashPtr == "" {
		log.Crit("No tx hash provided")
	}

	client, err := ethclient.Dial(*ethNodeUriPtr)
	if err != nil {
		log.Crit("Failed to connect to eth node", "err", err)
	}

	status, _, uncleBlock, err := txinfo.WasTxUncled(client, common.HexToHash(*txHashPtr))
	if err != nil {
		log.Crit("Failed to check tx", "err", err)
	}

	if status == txinfo.StatusTxUnknown {
		fmt.Println("tx not found")
	} else if status == txinfo.StatusTxNotUncled {
		fmt.Println("tx not uncled")
	} else if status == txinfo.StatusTxWasUncled {
		fmt.Printf("tx was uncled in block %s %s\n", uncleBlock.Number(), uncleBlock.Hash().Hex())
	}
}
