package testing

import (
	"fmt"
	"testing"
	"time"
)

var listUsers []string

func init() {
	for i := 1; i < 14; i++ {
		listUsers = append(listUsers, fmt.Sprintf("rol-user%d", i))
	}
}

func TestPerfsSingle(t *testing.T) {
	ps := NewPerfs(1000, nil)
	_, err := ps.runSingleUser(listUsers[0])
	if err != nil {
		t.Fatal(err)
	}
}

func TestPerfsAll(t *testing.T) {
	ps := NewPerfs(1000, nil)
	ps.Run()
}

func TestBet(t *testing.T) {
	testUser := listUsers[3]
	c, err := newPerfClient(testUser)
	if err != nil {
		t.Fatal(err)
	}

	_, seq, err := c.GetAccountNumberSequence()
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		res, err := c.sendTx(testUser, seq)
		if err != nil {
			t.Fatal(err)
		}
		if res.Code == 0 {
			seq += 1
		} else {
			time.Sleep(20 * time.Millisecond)
			_, seq, err = c.GetAccountNumberSequence()
			if err != nil {
				t.Fatal(err)
			}
		}
	}

}
