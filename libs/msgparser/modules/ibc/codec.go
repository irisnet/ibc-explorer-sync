package ibc

import (
	"github.com/irisnet/ibc-explorer-sync/libs/msgparser/codec"
	ibctransfer "github.com/okex/exchain/libs/ibc-go/modules/apps/transfer"
	ibc "github.com/okex/exchain/libs/ibc-go/modules/core"
)

func init() {
	codec.RegisterAppModules(
		ibc.AppModuleBasic{},
		ibctransfer.AppModuleBasic{},
	)
}
