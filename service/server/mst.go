package server

import (
	"context"
	"fmt"

	"github.com/vulturedb/vulture/service/rpc"
)

type MSTServer struct {
}

func (s *MSTServer) GetNode(ctx context.Context, in *rpc.GetNodeRequest) (*rpc.MSTNode, error) {
	fmt.Printf("%v\n", in.GetNodeHash())
	return &rpc.MSTNode{
		Level:    69,
		Low:      nil,
		Children: nil,
	}, nil
}
