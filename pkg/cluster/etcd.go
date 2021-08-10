package cluster

import (
	"context"
	"go-ms/utils"
	"log"
	"strings"
	"sync"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"
)

var (
	Etcd        *clientV3.Client
	Servers     = make(map[string][]string)
	ServersLock sync.Mutex
	ProjectName = "go-ms"
)

func EtcdRegister(etcdAddr, rpcPort, httpPort, serverName string) error {
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

	return ServerRegister(rpcPort, httpPort, serverName)
}

func ServerRegister(rpcPort, httpPort, serverName string) error {
	intranetIp := utils.GetIntranetIp()
	intranetRpcAddr := intranetIp + ":" + rpcPort
	intranetHttpAddr := intranetIp + ":" + httpPort
	//新建租约
	resp, err := Etcd.Grant(context.TODO(), 2)
	if err != nil {
		return err
	}
	//授予租约
	if err != nil {
		return err
	}
	key := ProjectName + "_" + serverName + "_" + intranetRpcAddr + "_" + intranetHttpAddr
	_, err = Etcd.Put(context.TODO(), key, "0", clientV3.WithLease(resp.ID))
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

	ServersLock.Lock()
	servers := GetAllServers()
	for _, server := range servers {
		arr := strings.Split(server, "_")
		serverName := arr[0]
		serverRpcAddr := arr[1]
		serverHttpAddr := arr[2]
		Servers[serverName] = append(Servers[serverName], serverRpcAddr+"_"+serverHttpAddr)
	}
	ServersLock.Unlock()

	go ServerWatcher()

	return nil
}

func ServerWatcher() {
	rch := Etcd.Watch(context.Background(), ProjectName+"_", clientV3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			arr := strings.Split(string(ev.Kv.Key), "_")
			serverName := arr[1]
			rpcAddr := arr[2]
			httpAddr := arr[3]
			serverAddr := rpcAddr + "_" + httpAddr
			switch ev.Type {
			case 0: //put
				AddServer(serverName, serverAddr)
				log.Printf("[%s] node [%s] join \n", serverName, serverAddr)
			case 1: //delete
				DelServer(serverName, serverAddr)
				log.Printf("[%s] node [%s] leave \n", serverName, serverAddr)
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
