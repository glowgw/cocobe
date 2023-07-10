package testing

import (
	"github.com/dymensionxyz/rollapp/app"
	"github.com/dymensionxyz/rollapp/app/params"
)

type testOpts struct {
	connections int
	clients     int
}

type perfClient struct {
	opts *testOpts

	ec params.EncodingConfig
}

func newPerfClient(opts *testOpts) *perfClient {
	return &perfClient{
		opts: opts,
		ec:   app.MakeEncodingConfig(),
	}
}

func (pc *perfClient) sendTx() error {
	// argOption := "over"
	// numberBetting := uint32(40)
	// coin, err := sdk.ParseCoinNormalized("20urax")
	// if err != nil {
	// 	return err
	// }
	//
	// txBuilder := pc.ec.TxConfig.NewTxBuilder()
	//
	// priv1, _, addr1 := keys.KeyPubAddrFromSecret([]byte("client1"))
	//
	// msg := types.NewMsgDiceBetting(
	// 	addr1.String(),
	// 	argOption,
	// 	numberBetting,
	// 	&coin,
	// )
	//
	// if err := msg.ValidateBasic(); err != nil {
	// 	return err
	// }
	return nil
}
