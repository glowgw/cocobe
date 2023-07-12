package main

import (
	"flag"
	"github.com/glowgw/cocobe/internal/testing"
)

var (
	batchSize uint64
	from      int
	to        int
)

func main() {
	flag.Uint64Var(&batchSize, "batch", 1, "batch size")
	flag.IntVar(&from, "from", 1, "from")
	flag.IntVar(&to, "to", 1, "to")

	flag.Parse()

	listUsers := testing.MakeListUsers(from, to)

	ps := testing.NewPerfs(batchSize, listUsers)
	ps.Run()
}
