package client

import (
	"github.com/irisnet/ibc-explorer-sync/libs/msgparser/codec"
	"github.com/irisnet/ibc-explorer-sync/libs/msgparser/modules/ibc"
	_ "github.com/irisnet/ibc-explorer-sync/libs/msgparser/modules/stride_module"
	_ "github.com/irisnet/ibc-explorer-sync/libs/msgparser/modules/uptick_module"
)

type MsgClient struct {
	Ibc ibc.Client
}

func NewMsgClient() MsgClient {
	codec.MakeEncodingConfig()
	return MsgClient{
		Ibc: ibc.NewClient(),
	}
}
