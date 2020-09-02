package main

import (
	"context"
	"fmt"

	cbor "github.com/ipfs/go-ipld-cbor"
	mh "github.com/multiformats/go-multihash"
	"github.com/wojtechnology/cado/server"
)

type MyStruct struct {
	Items map[string]MyStruct
	Foo   string
	Bar   []byte
	Baz   []int
}

func testStruct() MyStruct {
	return MyStruct{
		Items: map[string]MyStruct{
			"Foo": {
				Foo: "Foo",
				Bar: []byte("Bar"),
				Baz: []int{1, 2, 3, 4},
			},
			"Bar": {
				Bar: []byte("Bar"),
				Baz: []int{1, 2, 3, 4},
			},
		},
		Baz: []int{5, 1, 2},
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ipfs, err := server.SpawnDefault(ctx)
	if err != nil {
		fmt.Println("No IPFS repo available on the default path")
	}
	fmt.Println("IPFS node is running")

	cbor.RegisterCborType(MyStruct{})
	nd, err := cbor.WrapObject(testStruct(), mh.SHA2_256, -1)
	if err != nil {
		panic(fmt.Errorf("Could not wrap object: %s", err))
	}

	dagService := ipfs.Dag()
	err = dagService.Add(ctx, nd)
	if err != nil {
		panic(fmt.Errorf("Could not add node to dag: %s", err))
	}

	fmt.Printf("Added node %s\n", nd.Cid().String())

	nd2, err := dagService.Get(ctx, nd.Cid())
	if err != nil {
		panic(fmt.Errorf("Could not get node from dag: %s", err))
	}

	fmt.Printf("Got node %s\n", nd2.Cid().String())
	for _, ln := range nd2.Tree("", -1) {
		fmt.Println(ln)
	}
}
