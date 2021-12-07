package txinfo

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

type TxStatus string

var (
	StatusTxUnknown   TxStatus = "TxUnknown"
	StatusTxNotUncled TxStatus = "TxNotUncled"
	StatusTxWasUncled TxStatus = "TxWasUncled"
)

func WasTxUncled(client *ethclient.Client, txHash common.Hash) (status TxStatus, foundInUncleBlock *types.Block, err error) {
	tx, _, err := client.TransactionByHash(context.Background(), txHash)
	if err == ethereum.NotFound {
		return StatusTxUnknown, nil, nil
	} else if err != nil {
		return StatusTxUnknown, nil, errors.Wrap(err, "failed to get transaction by hash")
	}

	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return StatusTxUnknown, nil, errors.Wrap(err, "failed to get transaction receipt")
	}

	currentBlock, err := client.BlockByHash(context.Background(), receipt.BlockHash)
	if err != nil {
		return StatusTxUnknown, nil, errors.Wrap(err, "failed to get block by hash")
	}

	// check uncles of included block, and if not found check the previous few blocks for uncles
	for blocksTried := 0; blocksTried < 6; blocksTried++ {
		found, foundInBlock := IsTxFoundInOneOfBlockUncles(client, currentBlock, tx.Hash())
		if found {
			return StatusTxWasUncled, foundInBlock, nil
		}

		prevBlockNumber := big.NewInt(currentBlock.Number().Int64() - 1)
		currentBlock, err = client.BlockByNumber(context.Background(), prevBlockNumber)
		if err != nil {
			return StatusTxUnknown, nil, errors.Wrap(err, "failed to get block by number")
		}
	}

	return StatusTxNotUncled, nil, nil
}

func IsTxFoundInOneOfBlockUncles(client *ethclient.Client, block *types.Block, txHash common.Hash) (found bool, foundInBlock *types.Block) {
	for _, uncleHeader := range block.Uncles() {
		uncleBlock, err := client.BlockByHash(context.Background(), uncleHeader.Hash())
		if err != nil {
			log.Println("Failed to get uncle block", uncleHeader.Hash().Hex())
			continue
		}

		for _, uncleTx := range uncleBlock.Transactions() {
			if txHash == uncleTx.Hash() {
				return true, uncleBlock
			}
		}
	}
	return false, nil
}
