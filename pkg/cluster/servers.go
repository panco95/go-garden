package cluster

import (
	"context"
	"go-ms/pkg/base"
	clientV3 "go.etcd.io/etcd/client/v3"
	"strings"
	"time"
)

func GetAllServers() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := Etcd.Get(ctx, "go-ms_", clientV3.WithPrefix())
	cancel()
	if err != nil {
		base.Logger.Debugf(err.Error())
		return []string{}
	}
	var servers []string
	for _, ev := range resp.Kvs {
		arr := strings.Split(string(ev.Key), "go-ms_")
		server := arr[1]
		servers = append(servers, server)
	}
	return servers
}

func GetServersByName(serverName string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := Etcd.Get(ctx, "go-ms_"+serverName, clientV3.WithPrefix())
	cancel()
	if err != nil {
		base.Logger.Debugf(err.Error())
		return []string{}
	}
	var servers []string
	for _, ev := range resp.Kvs {
		arr := strings.Split(string(ev.Key), "go-ms_"+serverName+"_")
		serverAddr := arr[1]
		servers = append(servers, serverAddr)
	}
	return servers
}

func AddServer(serverName, serverAddr string) {
	ServersLock.Lock()
	Servers[serverName] = append(Servers[serverName], serverAddr)
	ServersLock.Unlock()
}

func DelServer(serverName, serverAddr string) {
	ServersLock.Lock()
	for i := 0; i < len(Servers[serverName]); i++ {
		if Servers[serverName][i] == serverAddr {
			Servers[serverName] = append(Servers[serverName][:i], Servers[serverName][i+1:]...)
			i--
		}
	}
	ServersLock.Unlock()
}
