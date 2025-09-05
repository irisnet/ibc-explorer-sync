package stride_module

import (
	"github.com/Stride-Labs/stride/v21/x/interchainquery"
	"github.com/irisnet/ibc-explorer-sync/libs/msgparser/codec"
)

func init() {
	codec.RegisterAppModules(
		interchainquery.AppModuleBasic{},
	)
}
