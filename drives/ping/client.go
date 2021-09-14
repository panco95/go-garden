package ping

import (
	"context"
	"google.golang.org/grpc"
)

// Ping ping rpc服务端
func Ping(address string) (string, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return "", err
	}
	defer conn.Close()

	c := NewPingClient(conn)

	req := &PingRequest{Msg: "ok"}
	res, err := c.Ping(context.Background(), req)

	if err != nil {
		return "", err
	}

	return res.Msg, nil
}
