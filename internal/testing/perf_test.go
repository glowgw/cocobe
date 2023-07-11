package testing

import (
	"fmt"
	"testing"
)

var listUsers []string

func init() {
	for i := 1; i < 8; i++ {
		listUsers = append(listUsers, fmt.Sprintf("rol-user%d", i))
	}
}

func TestBet(t *testing.T) {
	c, err := newPerfClient()
	if err != nil {
		t.Fatal(err)
	}
	iter := 0
	for {
		if iter > 1_000_000 {
			break
		}

		var seq uint64
		firstSeq, err := c.GetAccountNumberSequence(listUsers[0])
		if err != nil {
			t.Fatal(err)
		}
		seq = firstSeq
		accSeq, err := c.sendTx(listUsers[0], seq)
		if err != nil {
			t.Fatal(err)
		}
		seq = accSeq

		iter += 1
	}

}
