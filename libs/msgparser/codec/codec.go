package codec

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	ctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	injective "github.com/InjectiveLabs/sdk-go/chain/types"
	injectivecodec "github.com/InjectiveLabs/sdk-go/chain/crypto/codec"
	"github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

var (
	appModules []module.AppModuleBasic
	encodecfg  EncodingConfig
)

type EncodingConfig struct {
	InterfaceRegistry ctypes.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

// 初始化账户地址前缀
func MakeEncodingConfig() {
	var cdc = codec.NewLegacyAmino()
	cryptocodec.RegisterCrypto(cdc)

	interfaceRegistry := ctypes.NewInterfaceRegistry()
	moduleBasics := module.NewBasicManager(appModules...)
	moduleBasics.RegisterInterfaces(interfaceRegistry)
	std.RegisterInterfaces(interfaceRegistry)
	injective.RegisterInterfaces(interfaceRegistry)
	injectivecodec.RegisterInterfaces(interfaceRegistry)
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(marshaler, tx.DefaultSignModes)

	encodecfg = EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             marshaler,
		TxConfig:          txCfg,
		Amino:             cdc,
	}
}

func GetTxDecoder() sdk.TxDecoder {
	return encodecfg.TxConfig.TxDecoder()
}

func GetMarshaler() codec.Codec {
	return encodecfg.Codec
}

func GetSigningTx(txBytes types.Tx) (signing.Tx, error) {
	Tx, err := GetTxDecoder()(txBytes)
	if err != nil {
		return nil, err
	}
	signingTx := Tx.(signing.Tx)
	return signingTx, nil
}

func RegisterAppModules(module ...module.AppModuleBasic) {
	appModules = append(appModules, module...)
}
