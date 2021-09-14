package ping

import (
	"context"
)

type service struct{}

var Service = service{}

func (s service) Ping(ctx context.Context, in *PingRequest) (*PingResponse, error) {
	resp := new(PingResponse)
	resp.Msg = "ok"
	return resp, nil
}
