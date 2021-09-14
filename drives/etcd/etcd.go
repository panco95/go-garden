package etcd

import (
	"context"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"
)

var client *clientV3.Client

// GetClient 获取etcd客户端
func GetClient() *clientV3.Client {
	return client
}

// Connect 连接etcd
func Connect(etcdAddr []string) error {
	var err error
	client, err = clientV3.New(clientV3.Config{
		Endpoints:   etcdAddr,
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for _, addr := range etcdAddr {
		_, err = client.Status(timeoutCtx, addr)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetKV 获取etcd某个key的value
func GetKV(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	getResp, err := client.Get(ctx, key)
	cancel()
	if err != nil {
		return "", err
	}
	result := ""
	for _, val := range getResp.Kvs {
		result = string(val.Value)
	}
	return result, nil
}
