package keys

import "testing"

func TestKeysFromSecret(t *testing.T) {
	priv1, pub1, addr1 := KeyPubAddrFromSecret([]byte("client1"))
	t.Logf("priv1=%v, pub1=%v, addr1=%v", priv1, pub1, addr1.String())
}
