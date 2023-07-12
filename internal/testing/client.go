package testing

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
	kr keyring.Keyring

	user   string // from name
	accNum uint64

	baseCtx client.Context
	txf     tx.Factory
}

const (
	chainID = "rollapp"
)

var (
	gas    = uint64(1_000_000)
	gasAdj = 1.15
)

func newPerfClient(user string) (*perfClient, error) {
	logger := slog.Default()
	cfg := config.Default()

	nodeClient, err := client.NewClientFromNode(cfg.RpcUrl)
	if err != nil {
		return nil, err
	}

	ec := app.MakeEncodingConfig()

	baseCtx := client.Context{}.
		WithFromName(user).
		WithChainID(chainID).
		WithTxConfig(ec.TxConfig).
		WithCodec(ec.Codec).
		WithHomeDir(app.DefaultNodeHome).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithInterfaceRegistry(ec.InterfaceRegistry).
		WithClient(nodeClient)

	kr, err := keyring.New(sdk.KeyringServiceName(), "test", baseCtx.HomeDir, baseCtx.Input, baseCtx.Codec, baseCtx.KeyringOptions...)
	if err != nil {
		logger.Error("new keyring error", "err", err)
		return nil, err
	}

	txf := tx.Factory{}.
		WithChainID(chainID).
		WithTxConfig(ec.TxConfig).
		WithKeybase(kr).
		WithGas(gas).
		WithGasAdjustment(gasAdj).
		WithAccountRetriever(authtypes.AccountRetriever{})

	pc := &perfClient{
		ec:      ec,
		logger:  logger,
		baseCtx: baseCtx,
		txf:     txf,
		cfg:     cfg,
		kr:      kr,
		user:    user,
	}

	addr, err := pc.GetAddress()
	if err != nil {
		return nil, err
	}
	pc.baseCtx = pc.baseCtx.WithFromAddress(addr)

	accNum, seq, err := pc.GetAccountNumberSequence()
	if err != nil {
		return nil, err
	}
	pc.txf = pc.txf.WithAccountNumber(accNum).WithSequence(seq)

	return pc, nil
}

func (pc *perfClient) GetAccountNumberSequence() (uint64, uint64, error) {
	addr, err := pc.GetAddress()
	if err != nil {
		return 0, 0, err
	}
	accNum, accSeq, err := pc.txf.AccountRetriever().GetAccountNumberSequence(pc.baseCtx, addr)
	return accNum, accSeq, err
}

func (pc *perfClient) GetAddress() (sdk.AccAddress, error) {
	record, err := pc.kr.Key(pc.baseCtx.GetFromName())
	if err != nil {
		pc.logger.Error("err get key by from name", "err", err)
		return nil, err
	}
	addr, err := record.GetAddress()
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (pc *perfClient) sendTx(user string, seq uint64) (*sdk.TxResponse, error) {
	// Update account sequence each time send tx
	pc.txf = pc.txf.WithSequence(seq)

	argOption := "over"
	numberBetting := uint32(40)
	coin, err := sdk.ParseCoinNormalized("200urax")
	if err != nil {
		return nil, err
	}

	msg := types.NewMsgDiceBetting(
		pc.baseCtx.GetFromAddress().String(),
		argOption,
		numberBetting,
		&coin,
	)
	if err := msg.ValidateBasic(); err != nil {
		pc.logger.Error("validate basic failed", "err", err)
		return nil, err
	}

	pc.baseCtx.BroadcastMode = flags.BroadcastSync

	txUnsign, err := pc.txf.BuildUnsignedTx(msg)
	if err != nil {
		return nil, err
	}
	err = tx.Sign(pc.txf, user, txUnsign, true)
	if err != nil {
		pc.logger.Error("err sign: ", "err", err)
		return nil, err
	}

	txBytes, err := pc.baseCtx.TxConfig.TxEncoder()(txUnsign.GetTx())
	if err != nil {
		pc.logger.Error("encode tx error: ", "err", err)
		return nil, err
	}

	res, err := pc.baseCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}
	return res, err
}
