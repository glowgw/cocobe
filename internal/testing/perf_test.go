package testing

import (
	"testing"
)

func TestBet(t *testing.T) {
	c, err := newPerfClient()
	if err != nil {
		t.Fatal(err)
	}
	err = c.sendTx()
	if err != nil {
		t.Fatal(err)
	}

}
