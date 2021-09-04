package goms

import (
	"context"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"
)

var (
	Etcd *clientV3.Client
)

func EtcdConnect(etcdAddr []string) error {
	var err error
	Etcd, err = clientV3.New(clientV3.Config{
		Endpoints:   etcdAddr,
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for _, addr := range etcdAddr {
		_, err = Etcd.Status(timeoutCtx, addr)
		if err != nil {
			return err
		}
	}
	return nil
}


func GetKV(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	getResp, err := Etcd.Get(ctx, key)
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
