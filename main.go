package main

import (
	"context"
	"fmt"
	"time"

	"github.com/wojtechnology/cado/core"
	"github.com/wojtechnology/cado/server"
)

func exampleSchema() core.Schema {
	return core.Schema{Fields: map[string]core.FieldSpec{
		"id":       {Type: "int", Nullable: false},
		"username": {Type: "string", Nullable: false},
		"email":    {Type: "string", Nullable: true},
	}}
}

func validateRows() {
	s := exampleSchema()
	r := core.Row{
		Data: map[string]interface{}{
			"id":       int32(123),
			"username": "wojtechnology",
		},
	}
	err := s.ValidateRow(r)
	if err != nil {
		panic(fmt.Errorf("Error when validation row 1: %s", err))
	}

	r1 := core.Row{
		Data: map[string]interface{}{
			"id":       int32(123),
			"username": "wojtechnology",
		},
	}
	err = s.ValidateRow(r1)
	if err != nil {
		panic(fmt.Errorf("Error when validation row 1: %s", err))
	}

	r2 := core.Row{
		Data: map[string]interface{}{
			"id":       int32(123),
			"username": "wojtechnology",
			"email":    true,
		},
	}
	err = s.ValidateRow(r2)
	if err != nil {
		panic(fmt.Errorf("Error when validation row 2: %s", err))
	}
}

func serverExample() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server.RegisterTypes()
	ipfs, err := server.SpawnDefault(ctx)
	if err != nil {
		fmt.Println("No IPFS repo available on the default path")
	}
	fmt.Println("IPFS node is running")

	s := exampleSchema()
	cid, err := server.PutSchema(ctx, ipfs.Dag(), s)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Printf("Saved as %s\n", cid)

	start := time.Now()
	nd2, err := server.GetSchema(ctx, ipfs.Dag(), cid)
	t := time.Now()
	elapsed := t.Sub(start)
	if err != nil {
		panic(fmt.Errorf("Could not get node from dag: %s", err))
	}
	fmt.Printf("Got %v in %s\n", nd2, elapsed)
}

type Float float32

func (f Float) Less(than core.IndexElem) bool {
	return f < than.(Float)
}

func bTreeExample() {
	start := time.Now()
	index := core.NewBTreeIndex(25)
	for i := 0; i < 100; i++ {
		index.Put(Float(float32(i)))
		fmt.Printf("Step %d\n", i)
	}
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("Ran in %s\n", elapsed)
	index.PrintInOrder()
	index.Remove(Float(99.0))
	fmt.Printf("Removed\n")
	index.PrintInOrder()
}

func main() {
	bTreeExample()
}
