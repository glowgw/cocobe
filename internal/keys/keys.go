package keys

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/rollapp/app"
	"github.com/dymensionxyz/rollapp/utils"
)

func init() {
	sdkconfig := sdk.GetConfig()
	utils.SetPrefixes(sdkconfig, app.AccountAddressPrefix)
	utils.SetBip44CoinType(sdkconfig)
}

func KeyPubAddrFromSecret(secret []byte) (cryptotypes.PrivKey, cryptotypes.PubKey, sdk.AccAddress) {
	key := secp256k1.GenPrivKeyFromSecret(secret)
	pub := key.PubKey()
	acc := sdk.AccAddress(pub.Address())
	addr, err := sdk.AccAddressFromBech32(acc.String())
	if err != nil {
		panic(err)
	}
	return key, pub, addr
}
