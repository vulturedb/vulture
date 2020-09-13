package main

import (
	"context"
	"crypto"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/wojtechnology/cado/core"
	"github.com/wojtechnology/cado/index"
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

type Float uint32

func (f Float) Less(than index.Key) bool {
	return f < than.(Float)
}

func (f Float) PutBytes(w io.Writer) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(f))
	_, err := w.Write(buf)
	return err
}

func mstExample() {
	ind := index.NewLocalMST(index.Base16, crypto.SHA256)
	ind.Put(Float(420.0), Float(69.0))
	val := ind.Get(Float(420.0))
	fmt.Printf("%d\n", val)
	val = ind.Get(Float(123.0))
	fmt.Printf("%v\n", val)
	val = ind.Get(Float(1230.0))
	fmt.Printf("%v\n", val)
}

func main() {
	mstExample()
}
