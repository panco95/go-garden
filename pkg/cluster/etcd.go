package cluster

import (
	"context"
	"fmt"
	"go-ms/utils"
	"strings"
	"sync"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"
)

var (
	Etcd        *clientV3.Client
	Servers     = make(map[string][]string)
	ServersLock sync.Mutex
)

func EtcdRegister(etcdAddr string, rpcPort string, serverName string) error {
	addrArr := strings.Split(etcdAddr, "|")
	var err error
	Etcd, err = clientV3.New(clientV3.Config{
		Endpoints:   addrArr,
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for _, addr := range addrArr {
		_, err = Etcd.Status(timeoutCtx, addr)
		if err != nil {
			return err
		}
	}

	return ServerRegister(rpcPort, serverName)
}

func ServerRegister(rpcPort string, serverName string) error {
	intranetIp := utils.GetIntranetIp()
	intranetRpcAddr := intranetIp + ":" + rpcPort
	//新建租约
	resp, err := Etcd.Grant(context.TODO(), 5)
	if err != nil {
		return err
	}
	//授予租约
	if err != nil {
		return err
	}
	key := serverName + "_" + intranetRpcAddr
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
	Servers[serverName] = GetServers(serverName)
	ServersLock.Unlock()

	go serverWatcher(serverName)

	return nil
}

func serverWatcher(serverName string) {
	rch := Etcd.Watch(context.Background(), serverName+"_", clientV3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			arr := strings.Split(string(ev.Kv.Key), serverName+"_")
			serverAddr := arr[1]
			switch ev.Type {
			case 0: //put
				AddServer(serverName, serverAddr)
				fmt.Printf("Server [%s] cluster [%s] join \n", serverName, serverAddr)
			case 1: //delete
				DelServer(serverName, serverAddr)
				fmt.Printf("Server [%s] cluster [%s] leave \n", serverName, serverAddr)
			}
		}
	}
}