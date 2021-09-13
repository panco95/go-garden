package redis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
)

var client *redis.Client

// Connect 连接redis
func Connect(address string) error {
	client = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})
	if client.Ping(context.Background()).Err() != nil {
		return errors.New("Connect error")
	}
	return nil
}

// Client 获取redis客户端
func Client() *redis.Client {
	return client
}
