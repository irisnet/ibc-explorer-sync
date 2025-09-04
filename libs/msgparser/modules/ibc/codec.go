package ibc

import (
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclients "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/irisnet/ibc-explorer-sync/libs/msgparser/codec"
)

func init() {
	codec.RegisterAppModules(
		ibc.AppModuleBasic{},
		ibctransfer.AppModuleBasic{},
		ibcclients.AppModuleBasic{},
	)
}
