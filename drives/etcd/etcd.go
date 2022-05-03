package etcd

import (
	"context"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

func Connect(addr []string, logger *zap.Logger) (*clientV3.Client, error) {
	etcd, err := clientV3.New(clientV3.Config{
		Endpoints:   addr,
		DialTimeout: 3 * time.Second,
		Logger:      logger,
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
