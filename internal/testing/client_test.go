package testing

import (
	"fmt"
	"testing"
	"time"
)

var listUsers []string

func init() {
	for i := 1; i < 8; i++ {
		listUsers = append(listUsers, fmt.Sprintf("rol-user%d", i))
	}
}

func TestPerfsSingle(t *testing.T) {
	ps := newPerfs(1000)
	_, err := ps.runSingleUser(listUsers[0])
	if err != nil {
		t.Fatal(err)
	}
}

func TestPerfsAll(t *testing.T) {
	ps := newPerfs(1000)
	ps.run()
}

func TestBet(t *testing.T) {
	c, err := newPerfClient()
	if err != nil {
		t.Fatal(err)
	}

	seq, err := c.GetAccountNumberSequence(listUsers[0])
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		res, err := c.sendTx(listUsers[0], seq)
		if err != nil {
			t.Fatal(err)
		}
		if res.Code == 0 {
			seq += 1
		} else {
			time.Sleep(20 * time.Millisecond)
			seq, err = c.GetAccountNumberSequence(listUsers[0])
			if err != nil {
				t.Fatal(err)
			}
		}
	}

}