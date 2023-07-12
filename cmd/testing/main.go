package main

import "github.com/glowgw/cocobe/internal/testing"

func main() {
	ps := testing.NewPerfs(10000)
	ps.Run()
}
