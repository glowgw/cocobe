package testing

import (
	"fmt"
	"golang.org/x/exp/slog"
	"sync"
	"time"
)

type perfs struct {
	batchSize  uint64
	numClients uint64
	clients    map[string]*perfClient // name to client
	listUsers  []string

	logger *slog.Logger
}

func NewPerfs(batchSize uint64) *perfs {
	listUsers := makeListUsers()
	clients := map[string]*perfClient{}
	for _, name := range listUsers {
		c, err := newPerfClient()
		if err != nil {
			panic(err)
		}
		clients[name] = c
	}
	return &perfs{
		batchSize: batchSize,
		clients:   clients,
		listUsers: listUsers,
		logger:    slog.Default(),
	}
}

func makeListUsers() []string {
	var listUsers []string
	for i := 1; i < 14; i++ {
		listUsers = append(listUsers, fmt.Sprintf("rol-user%d", i))
	}
	return listUsers
}

func (p *perfs) runSingleUser(name string) (uint64, error) {
	c := p.clients[name]
	seq, err := c.GetAccountNumberSequence(name)
	if err != nil {
		p.logger.Error("", "err", err)
		return 0, err
	}

	total := uint64(0)

	for i := 0; uint64(i) < p.batchSize; i++ {
		res, err := c.sendTx(name, seq)
		if err != nil {
			return 0, err
		}

		if res.Code == 0 {
			seq++
			total++
		} else {
			time.Sleep(20 * time.Millisecond)
			seq, err = c.GetAccountNumberSequence(name)
			if err != nil {
				p.logger.Error("", "err", err)
				return 0, err
			}
		}

	}
	return total, nil
}

func (p *perfs) Run() {
	startTime := time.Now()

	var wg sync.WaitGroup
	wg.Add(len(p.listUsers))

	totals := make([]uint64, len(p.listUsers))

	for i, user := range p.listUsers {
		go func(user string, i int) {
			p.logger.Info("User sending tx", "user", user)
			defer wg.Done()
			total, err := p.runSingleUser(user)
			if err != nil {
				p.logger.Error("", "err", err)
			}
			p.logger.Info("User send done", "user", user)
			totals[i] = total
		}(user, i)
	}

	wg.Wait()
	diff := time.Now().Sub(startTime)
	totalMsgs := sum(totals)
	p.logger.Info("total time", "time", diff.Seconds())
	p.logger.Info("total message success", "message", totalMsgs)
	p.logger.Info("tps", "tps", float64(totalMsgs)/diff.Seconds())
}

func sum(inputs []uint64) uint64 {
	s := uint64(0)
	for _, input := range inputs {
		s += input
	}
	return s
}
