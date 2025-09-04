package handlers

import (
	"testing"

	"github.com/irisnet/ibc-explorer-sync/config"
	"github.com/irisnet/ibc-explorer-sync/libs/pool"
	"github.com/irisnet/ibc-explorer-sync/models"
	"github.com/irisnet/ibc-explorer-sync/utils"
)

func Test_handleTxResult(t *testing.T) {
	conf, err := config.ReadConfig()
	if err != nil {
		t.Fatal(err.Error())
	}
	models.Init(conf)
	InitRouter(conf)
	pool.Init(conf)
	c := pool.GetClient()
	defer func() {
		c.Release()
	}()
	b := int64(14121588)
	block, err := c.Block(&b)
	if err != nil {
		t.Fatal(err.Error())
	}
	res := handleTxResult(c, block.Block)
	for _, val := range res {
		t.Log(val.TxHash, utils.MarshalJsonIgnoreErr(val.TxResult))
	}
}
