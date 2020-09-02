package main

import (
	"context"
	"fmt"

	"github.com/wojtechnology/cado/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := server.SpawnDefault(ctx)
	if err != nil {
		fmt.Println("No IPFS repo available on the default path")
	}
	fmt.Println("IPFS node is running")
}
