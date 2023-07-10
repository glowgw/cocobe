package testing

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dymensionxyz/rollapp/app"
	"github.com/dymensionxyz/rollapp/app/params"
	"github.com/dymensionxyz/rollapp/utils"
	"github.com/dymensionxyz/rollapp/x/classicdice/types"
	"github.com/glowgw/cocobe/internal/config"
	"golang.org/x/exp/slog"
)

func init() {
	sdkconfig := sdk.GetConfig()
	utils.SetPrefixes(sdkconfig, app.AccountAddressPrefix)
	utils.SetBip44CoinType(sdkconfig)
	sdkconfig.Seal()
}

type perfClient struct {
	cfg    *config.Config
	logger *slog.Logger

	ec params.EncodingConfig

	baseCtx client.Context
	txf     tx.Factory
}

const (
	chainID = "rollapp"
)

var (
	username = "rol-user3"
	gas      = uint64(1_000_000)
	gasAdj   = 1.15
)

func newPerfClient() (*perfClient, error) {
	logger := slog.Default()
	cfg := config.Default()

	ec := app.MakeEncodingConfig()

	baseCtx := client.Context{}.
		WithChainID(chainID).
		WithTxConfig(ec.TxConfig).
		WithCodec(ec.Codec).
		WithHomeDir(app.DefaultNodeHome).
		WithAccountRetriever(types2.AccountRetriever{}).
		WithInterfaceRegistry(ec.InterfaceRegistry)
	nodeClient, err := client.NewClientFromNode(cfg.RpcUrl)
	if err != nil {
		return nil, err
	}
	baseCtx = baseCtx.WithClient(nodeClient)
	baseCtx = baseCtx.WithFromName(username)
	kr, err := keyring.New(sdk.KeyringServiceName(), "test", baseCtx.HomeDir, baseCtx.Input, baseCtx.Codec, baseCtx.KeyringOptions...)
	if err != nil {
		logger.Error("new keyring error", "err", err)
		return nil, err
	}

	record, err := kr.Key(baseCtx.GetFromName())
	if err != nil {
		logger.Error("err get key by from name", "err", err)
		return nil, err
	}
	addr, err := record.GetAddress()
	if err != nil {
		return nil, err
	}
	logger.Info("addr by from name", "addr", addr, "from name", baseCtx.FromName)
	baseCtx = baseCtx.WithFromAddress(addr)
	txf := tx.Factory{}.WithChainID(chainID).WithTxConfig(ec.TxConfig).WithKeybase(kr).WithGas(gas).WithGasAdjustment(gasAdj).WithAccountRetriever(types2.AccountRetriever{})

	fromAddr := baseCtx.GetFromAddress()
	logger.Info("fromAddr", "val", fromAddr)

	accNum, _, err := txf.AccountRetriever().GetAccountNumberSequence(baseCtx, fromAddr)
	if err != nil {
		logger.Error("account retriever failed", "err", err)
		return nil, err
	}
	txf = txf.WithAccountNumber(accNum)

	return &perfClient{
		ec:      ec,
		logger:  logger,
		baseCtx: baseCtx,
		txf:     txf,
		cfg:     cfg,
	}, nil
}

func (pc *perfClient) sendTx() error {
	argOption := "over"
	numberBetting := uint32(40)
	coin, err := sdk.ParseCoinNormalized("200urax")
	if err != nil {
		return err
	}

	msg := types.NewMsgDiceBetting(
		pc.baseCtx.GetFromAddress().String(),
		argOption,
		numberBetting,
		&coin,
	)
	if err := msg.ValidateBasic(); err != nil {
		pc.logger.Error("validate basic failed", "err", err)
		return err
	}

	pc.baseCtx.BroadcastMode = flags.BroadcastSync

	txUnsign, err := pc.txf.BuildUnsignedTx(msg)
	if err != nil {
		return err
	}
	err = tx.Sign(pc.txf, username, txUnsign, true)
	if err != nil {
		pc.logger.Error("err sign: ", "err", err)
		return err
	}

	txBytes, err := pc.baseCtx.TxConfig.TxEncoder()(txUnsign.GetTx())
	if err != nil {
		pc.logger.Error("encode tx error: ", "err", err)
		return err
	}

	res, err := pc.baseCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	pc.logger.Info("broadcast tx: ", "code", res.Code, "tx_hash", res.TxHash)

	return nil
}
