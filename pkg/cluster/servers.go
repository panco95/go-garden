package cluster

import (
	"context"
	"errors"
	"go-ms/pkg/base/global"
	clientV3 "go.etcd.io/etcd/client/v3"
	"strings"
	"sync"
	"time"
)

type Server struct {
	PollNext      int
	Nodes         []string
	RequestFinish int
}

var (
	Servers     = make(map[string]*Server)
	ServersLock sync.Mutex
	ProjectName = "go-ms"
)

func GetAllServers() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := Etcd.Get(ctx, ProjectName+"_", clientV3.WithPrefix())
	cancel()
	if err != nil {
		global.Logger.Debugf(err.Error())
		return []string{}
	}
	var servers []string
	for _, ev := range resp.Kvs {
		arr := strings.Split(string(ev.Key), ProjectName+"_")
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
		global.Logger.Debugf(err.Error())
		return []string{}
	}
	var servers []string
	for _, ev := range resp.Kvs {
		arr := strings.Split(string(ev.Key), ProjectName+"_"+serverName+"_")
		serverAddr := arr[1]
		servers = append(servers, serverAddr)
	}
	return servers
}

func AddServer(serverName, serverAddr string) {
	ServersLock.Lock()
	ExistsServer(serverName)
	Servers[serverName].Nodes = append(Servers[serverName].Nodes, serverAddr)
	ServersLock.Unlock()
}

func DelServer(serverName, serverAddr string) {
	ServersLock.Lock()
	ExistsServer(serverName)
	for i := 0; i < len(Servers[serverName].Nodes); i++ {
		if Servers[serverName].Nodes[i] == serverAddr {
			Servers[serverName].Nodes = append(Servers[serverName].Nodes[:i], Servers[serverName].Nodes[i+1:]...)
			i--
		}
	}
	ServersLock.Unlock()
}

func ExistsServer(serverName string) {
	if _, ok := Servers[serverName]; !ok {
		Servers[serverName] = &Server{
			PollNext:      0,
			Nodes:         []string{},
			RequestFinish: 0,
		}
	}
}

func AnalyzeRpcAddr(serverName string, index int) (string, error) {
	if index > len(Servers[serverName].Nodes)-1 {
		return "", errors.New("Server not found")
	}
	arr := strings.Split(Servers[serverName].Nodes[index], "_")
	return arr[0], nil
}

func AnalyzeHttpAddr(serverName string, index int) (string, error) {
	if index > len(Servers[serverName].Nodes)-1 {
		return "", errors.New("Server not found")
	}
	arr := strings.Split(Servers[serverName].Nodes[index], "_")
	return arr[1], nil
}
