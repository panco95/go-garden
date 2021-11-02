package etcd

import (
	"context"
	clientV3 "go.etcd.io/etcd/client/v3"
	"time"
)

func Connect(addr []string) (*clientV3.Client, error) {
	etcd, err := clientV3.New(clientV3.Config{
		Endpoints:   addr,
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for _, addr := range addr {
		_, err = etcd.Status(timeoutCtx, addr)
		if err != nil {
			return nil, err
		}
	}
	return etcd, nil
}
