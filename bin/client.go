package main

import (
	"context"
	"fmt"

	"github.com/vulturedb/vulture/service/rpc"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:4000", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := rpc.NewMSTServiceClient(conn)
	nd, err := client.GetNode(context.Background(), &rpc.GetNodeRequest{
		NodeHash: []byte{1, 2, 3},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", *nd)
}
