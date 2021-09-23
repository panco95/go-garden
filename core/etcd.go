package core

import (
	"context"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"
)

func (g *Garden) connEtcd(etcdAddr []string) error {
	var err error
	g.etcd, err = clientV3.New(clientV3.Config{
		Endpoints:   etcdAddr,
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for _, addr := range etcdAddr {
		_, err = g.etcd.Status(timeoutCtx, addr)
		if err != nil {
			return err
		}
	}
	return nil
}
