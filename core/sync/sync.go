package sync

import (
	"context"
	"github.com/panco95/go-garden/core/utils"
	"google.golang.org/grpc"
)

type server struct{}

// Server grpc server
var Server = server{}

// SyncRoutes receive routes.yml and write file
func (s server) SyncRoutes(ctx context.Context, in *SyncRoutesRequest) (*SyncRoutesResponse, error) {
	resp := new(SyncRoutesResponse)
	resp.Result = true

	if err := utils.WriteFile("configs/routes.yml", in.Data); err != nil {
		resp.Result = false
	}

	return resp, nil
}

// SendSyncRoutes routes.yml to each other service
func SendSyncRoutes(address string, data []byte) (bool, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return false, err
	}
	defer conn.Close()

	c := NewSyncClient(conn)

	req := &SyncRoutesRequest{Data: data}
	res, err := c.SyncRoutes(context.Background(), req)

	if err != nil {
		return false, err
	}

	return res.Result, nil
}
