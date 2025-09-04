package ibc

import (
	"github.com/cosmos/cosmos-sdk/types"
	. "github.com/irisnet/ibc-explorer-sync/libs/msgparser/modules"
)

type Client interface {
	HandleTxMsg(v types.Msg) (MsgDocInfo, bool)
}
