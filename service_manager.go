package goms

import (
	"context"
	"errors"
	clientV3 "go.etcd.io/etcd/client/v3"
	"goms/utils"
	"strings"
	"sync"
	"time"
)

type service struct {
	PollNext      int
	Nodes         []string
	RequestFinish int
}

var (
	ProjectName  = "goms"
	Services     = make(map[string]*service)
	ServicesLock sync.Mutex
	ServiceId    string
	ServiceName  string
)

func InitServiceId(projectName, rpcPort, httpPort, serviceName string) {
	ServiceName = serviceName
	intranetIp := utils.GetOutboundIP()
	intranetRpcAddr := intranetIp + ":" + rpcPort
	intranetHttpAddr := intranetIp + ":" + httpPort
	ServiceId = projectName + "_" + ServiceName + "_" + intranetRpcAddr + "_" + intranetHttpAddr
}

func GetAllServices() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := Etcd.Get(ctx, ProjectName+"_", clientV3.WithPrefix())
	cancel()
	if err != nil {
		Logger.Debugf(err.Error())
		return []string{}
	}
	var services []string
	for _, ev := range resp.Kvs {
		arr := strings.Split(string(ev.Key), ProjectName+"_")
		service := arr[1]
		services = append(services, service)
	}
	return services
}

func GetServicesByName(serviceName string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := Etcd.Get(ctx, ProjectName+"_"+serviceName, clientV3.WithPrefix())
	cancel()
	if err != nil {
		Logger.Debugf(err.Error())
		return []string{}
	}
	var services []string
	for _, ev := range resp.Kvs {
		arr := strings.Split(string(ev.Key), ProjectName+"_"+serviceName+"_")
		serviceAddr := arr[1]
		services = append(services, serviceAddr)
	}
	return services
}

func AddService(name, addr string) {
	ServicesLock.Lock()
	ExistsService(name)
	Services[name].Nodes = append(Services[name].Nodes, addr)
	ServicesLock.Unlock()
}

func DelService(name, addr string) {
	ServicesLock.Lock()
	ExistsService(name)
	for i := 0; i < len(Services[name].Nodes); i++ {
		if Services[name].Nodes[i] == addr {
			Services[name].Nodes = append(Services[name].Nodes[:i], Services[name].Nodes[i+1:]...)
			i--
		}
	}
	ServicesLock.Unlock()
}

func ExistsService(name string) {
	if _, ok := Services[name]; !ok {
		Services[name] = &service{
			PollNext:      0,
			Nodes:         []string{},
			RequestFinish: 0,
		}
	}
}

func AnalyzeRpcAddr(name string, index int) (string, error) {
	if index > len(Services[name].Nodes)-1 {
		return "", errors.New("Service not found")
	}
	arr := strings.Split(Services[name].Nodes[index], "_")
	return arr[0], nil
}

func AnalyzeHttpAddr(name string, index int) (string, error) {
	if index > len(Services[name].Nodes)-1 {
		return "", errors.New("service node not found")
	}
	arr := strings.Split(Services[name].Nodes[index], "_")
	return arr[1], nil
}
