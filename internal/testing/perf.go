package testing

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/rollapp/app"
	"github.com/dymensionxyz/rollapp/app/params"
	"github.com/dymensionxyz/rollapp/x/classicdice/types"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slog"
)

type testOpts struct {
	connections int
	clients     int
}

type perfClient struct {
	opts   *testOpts
	logger *slog.Logger

	ec params.EncodingConfig

	baseCtx client.Context
	txf     tx.Factory
}

func newPerfClient(opts *testOpts) (*perfClient, error) {
	ec := app.MakeEncodingConfig()
	baseCtx := client.Context{}
	kr, err := client.NewKeyringFromBackend(baseCtx, "test")
	if err != nil {
		return nil, err
	}
	baseCtx = baseCtx.WithChainID("rollapp").WithTxConfig(ec.TxConfig).WithCodec(ec.Codec).WithHomeDir(app.DefaultNodeHome).WithKeyring(kr)
	txf := tx.NewFactoryCLI(baseCtx, &pflag.FlagSet{})

	return &perfClient{
		opts:    opts,
		ec:      ec,
		logger:  slog.Default(),
		baseCtx: baseCtx,
		txf:     txf,
	}, nil
}

func (pc *perfClient) sendTx() error {
	argOption := "over"
	numberBetting := uint32(40)
	coin, err := sdk.ParseCoinNormalized("20urax")
	if err != nil {
		return err
	}

	msg := types.NewMsgDiceBetting(
		"rol1pscmj2agc7d0h32g9v97ll6w35c9legx4u8tw5",
		argOption,
		numberBetting,
		&coin,
	)
	pc.baseCtx.BroadcastMode = flags.BroadcastSync

	txUnsign, err := pc.txf.BuildUnsignedTx(msg)
	if err != nil {
		return err
	}
	err = tx.Sign(pc.txf, "rol-user1", txUnsign, true)
	if err != nil {
		return err
	}

	txBytes, err := pc.baseCtx.TxConfig.TxEncoder()(txUnsign.GetTx())
	if err != nil {
		return err
	}

	res, err := pc.baseCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	pc.logger.Info("res", res)

	return nil
}
