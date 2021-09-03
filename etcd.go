package goms

import (
	"context"
	"log"
	"strings"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"
)

var (
	Etcd *clientV3.Client
)

func InitEtcd(etcdAddr string) error {
	addrArr := strings.Split(etcdAddr, "|")
	var err error
	Etcd, err = clientV3.New(clientV3.Config{
		Endpoints:   addrArr,
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for _, addr := range addrArr {
		_, err = Etcd.Status(timeoutCtx, addr)
		if err != nil {
			return err
		}
	}

	return ServiceRegister()
}

func ServiceRegister() error {
	//新建租约
	resp, err := Etcd.Grant(context.TODO(), 2)
	if err != nil {
		return err
	}
	//授予租约
	if err != nil {
		return err
	}
	_, err = Etcd.Put(context.TODO(), ServiceId, "0", clientV3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	//keep-alive
	ch, kaerr := Etcd.KeepAlive(context.TODO(), resp.ID)
	if kaerr != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ch:
				//keep-alive
			}
		}
	}()

	ServicesLock.Lock()
	services := GetAllServices()
	for _, service := range services {
		arr := strings.Split(service, "_")
		serviceName := arr[0]
		serviceRpcAddr := arr[1]
		serviceHttpAddr := arr[2]

		ExistsService(serviceName)
		Services[serviceName].Nodes = append(Services[serviceName].Nodes, serviceRpcAddr+"_"+serviceHttpAddr)
	}
	ServicesLock.Unlock()

	go ServiceWatcher()

	return nil
}

func ServiceWatcher() {
	rch := Etcd.Watch(context.Background(), ProjectName+"_", clientV3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			arr := strings.Split(string(ev.Kv.Key), "_")
			serviceName := arr[1]
			rpcAddr := arr[2]
			httpAddr := arr[3]
			serviceAddr := rpcAddr + "_" + httpAddr
			switch ev.Type {
			case 0: //put
				AddService(serviceName, serviceAddr)
				log.Printf("[%s] node [%s] join \n", serviceName, serviceAddr)
			case 1: //delete
				DelService(serviceName, serviceAddr)
				log.Printf("[%s] node [%s] leave \n", serviceName, serviceAddr)
			}
		}
	}
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
