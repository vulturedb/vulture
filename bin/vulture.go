package main

import (
	"fmt"
	"net"

	"github.com/vulturedb/vulture/service/rpc"
	"github.com/vulturedb/vulture/service/server"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 4000))
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	rpc.RegisterMSTServiceServer(grpcServer, &server.MSTServer{})
	err = grpcServer.Serve(lis)
	if err != nil {
		panic(err)
	}
}
