package main

import (
	"context"
	"fmt"

	"github.com/wojtechnology/cado/core"
	"github.com/wojtechnology/cado/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server.RegisterTypes()
	ipfs, err := server.SpawnDefault(ctx)
	if err != nil {
		fmt.Println("No IPFS repo available on the default path")
	}
	fmt.Println("IPFS node is running")

	s := core.Schema{Fields: map[string]core.FieldSpec{
		"id":       {Type: "int", Nullable: false},
		"username": {Type: "string", Nullable: false},
		"email":    {Type: "string", Nullable: true},
	}}
	cid, err := server.PutSchema(ctx, ipfs.Dag(), s)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Printf("Saved as %s\n", cid)

	nd2, err := server.GetSchema(ctx, ipfs.Dag(), cid)
	if err != nil {
		panic(fmt.Errorf("Could not get node from dag: %s", err))
	}
	fmt.Printf("Got %v\n", nd2)
}
