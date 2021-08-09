package client

import (
	"context"
	pb "go-ms/cmd/services/user/proto"
	"google.golang.org/grpc"
)

type UserService struct{}

func (h UserService) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	resp := new(pb.RegisterResponse)

	username := in.Username
	password := in.Password
	if username == "" || password == "" {
		resp.Result = false
		resp.Message = " Login Fail"
	} else {
		resp.Result = true
		resp.Message = " Login Success"
	}

	return resp, nil
}

func (h UserService) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	resp := new(pb.LoginResponse)
	resp.Result = false

	username := in.Username
	password := in.Password
	if username == "" || password == "" {
		resp.Result = false
		resp.Message = " Register Fail"
	} else {
		resp.Result = true
		resp.Message = "Register Success"
	}

	return resp, nil
}

func Call(rpcAddress string) (interface{}, error) {
	conn, err := grpc.Dial(rpcAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	c := pb.NewUserClient(conn)

	req := &pb.LoginRequest{Username: "Test", Password: "Test"}
	res, err := c.Login(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
