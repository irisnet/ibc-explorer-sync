package handlers

import (
	"context"
	"time"

	"github.com/irisnet/ibc-explorer-sync/libs/logger"
	"github.com/irisnet/ibc-explorer-sync/libs/msgparser/codec"
	"github.com/irisnet/ibc-explorer-sync/libs/pool"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

type chanTxResult struct {
	TxHash   string
	TxResult *ctypes.ResultTx
	Err      error
}

// parse tx with more goroutine concurrency
func handleTxResult(client *pool.Client, block *types.Block) map[string]chanTxResult {
	if _conf == nil {
		logger.Fatal("InitRouter don't work")
	}
	//control num of parse txReult goroutine concurrency
	if _conf.Server.WorkerNumTxResultInBlock <= 0 {
		_conf.Server.WorkerNumTxResultInBlock = 5
	}
	chanParseTxLimit := make(chan bool, _conf.Server.WorkerNumTxResultInBlock)
	chanRes := make(chan chanTxResult, len(block.Txs))
	for _, v := range block.Txs {
		chanParseTxLimit <- true
		// parse txReult with more goroutine concurrency
		go getTxResult(client, v, chanParseTxLimit, chanRes, block.Height)
	}
	txRetMap := make(map[string]chanTxResult, len(block.Txs))
	for i := 0; i < len(block.Txs); i++ {
		chanValue := <-chanRes
		txRetMap[chanValue.TxHash] = chanValue
	}
	return txRetMap
}
func includeIbcTxs(txBytes types.Tx, height int64) bool {
	var inclueIbcTx bool
	authTx, err := codec.GetSigningTx(txBytes)
	if err != nil {
		logger.Warn(err.Error(), logger.Int64("height", height))
		return inclueIbcTx
	}
	msgs := authTx.GetMsgs()
	if len(msgs) == 0 {
		return inclueIbcTx
	}
	for _, v := range msgs {
		_, ok := _filterMap[_parser.MsgType(v)]
		if ok {
			return true
		}
	}
	return inclueIbcTx
}

func getTxResult(c *pool.Client, txBytes types.Tx, chanLimit chan bool, chanRes chan chanTxResult, height int64) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("execute getTxResult fail", logger.Any("err", r), logger.Int64("height", height))
		}
		<-chanLimit
	}()
	var (
		txResult *ctypes.ResultTx
		err      error
	)
	if includeIbcTxs(txBytes, height) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		txResult, err = c.Tx(ctx, txBytes.Hash(), false)
		if err != nil {
			time.Sleep(1 * time.Second)
			if v, err1 := c.Tx(ctx, txBytes.Hash(), false); err1 != nil {
				err = err1
			} else {
				txResult = v
			}
		}
	}

	if txResult == nil {
		chanRes <- chanTxResult{Err: err}
		return
	}
	ret := chanTxResult{
		TxHash:   txResult.Hash.String(),
		TxResult: txResult,
		Err:      err,
	}
	chanRes <- ret

	return
}
