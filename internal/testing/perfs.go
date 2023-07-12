package testing

import "fmt"

type perfs struct {
	batchSize  uint64
	numClients uint64
	clients    map[string]*perfClient // name to client
	listUsers  []string
}

func newPerfs(batchSize, numClients uint64) *perfs {
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
		batchSize:  batchSize,
		numClients: numClients,
		clients:    clients,
		listUsers:  listUsers,
	}
}

func makeListUsers() []string {
	var listUsers []string
	for i := 1; i < 8; i++ {
		listUsers = append(listUsers, fmt.Sprintf("rol-user%d", i))
	}
	return listUsers
}

func (p *perfs) runSingleUser(name string) error {
	return nil
}

func (p *perfs) run() error {
	return nil
}
