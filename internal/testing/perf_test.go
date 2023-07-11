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
	var seq uint64
	iter := 0
	firstSeq, err := c.GetAccountNumberSequence(listUsers[0])
	if err != nil {
		t.Fatal(err)
	}
	seq = firstSeq
	t.Logf("seq = %d", firstSeq)
	for {
		if iter > 1_000_000 {
			break
		}
		accSeq, err := c.sendTx(listUsers[0], seq)
		if err != nil {
			t.Fatal(err)
		}
		seq = accSeq

		iter += 1
	}

}
