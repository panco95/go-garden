package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
)

func Connect(redisConfig map[string]interface{}, f func(err interface{})) (*redis.Client, error) {
	defer func() {
		if err := recover(); err != nil {
			f(err)
		}
	}()

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConfig["host"].(string) + ":" + redisConfig["port"].(string),
		Password: redisConfig["pass"].(string),
		DB:       redisConfig["db"].(int),
	})
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}
