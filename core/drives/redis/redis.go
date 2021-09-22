package redis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
)

var client *redis.Client

func Client() *redis.Client {
	return client
}

func Connect(address, password string, db int) error {
	client = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
	if client.Ping(context.Background()).Err() != nil {
		return errors.New("connect error")
	}
	return nil
}
