package client

import (
	"github.com/irisnet/ibc-explorer-sync/libs/msgparser/codec"
	"github.com/irisnet/ibc-explorer-sync/libs/msgparser/modules/ibc"
)

type MsgClient struct {
	Ibc ibc.Client
}

func NewMsgClient() MsgClient {
	codec.InitTxDecoder()
	return MsgClient{
		Ibc: ibc.NewClient(),
	}
}
