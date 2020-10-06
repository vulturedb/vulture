package main

import (
	"context"
	"crypto"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"

	mh "github.com/multiformats/go-multihash"
	"google.golang.org/grpc"

	"github.com/vulturedb/vulture/ipfs"
	"github.com/vulturedb/vulture/mst"
	"github.com/vulturedb/vulture/service/rpc"
	"github.com/vulturedb/vulture/service/server"
)

type UInt32KeyReader struct{}

func (kr UInt32KeyReader) FromBytes(b []byte) (mst.Key, error) {
	return mst.UInt32(binary.LittleEndian.Uint32(b)), nil
}

type UInt32ValueReader struct{}

func (kr UInt32ValueReader) FromBytes(b []byte) (mst.Value, error) {
	return mst.UInt32(binary.LittleEndian.Uint32(b)), nil
}

var host = flag.String("host", "localhost", "post to serve API for")
var port = flag.Int("port", 6667, "port to serve API on")

func main() {
	flag.Parse()

	// Create the MST and the MST Server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	coreAPI, err := ipfs.SpawnDefault(ctx)
	if err != nil {
		log.Fatalf("Failed to create ipfs node: %v", err)
	}
	log.Printf("IPFS node is running")
	ipfs.RegisterTypes()
	store := ipfs.NewIPFSMSTNodeStore(
		ctx,
		coreAPI.Dag(),
		mh.SHA2_256,
		UInt32KeyReader{},
		UInt32ValueReader{},
	)
	tree := mst.NewMST(mst.Base16, crypto.SHA256, store)
	mstServer := server.NewMSTServer(tree)

	// Start the grpc server
	address := fmt.Sprintf("%s:%d", *host, *port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Listening on %s", address)

	grpcServer := grpc.NewServer()
	rpc.RegisterMSTServiceServer(grpcServer, mstServer)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
