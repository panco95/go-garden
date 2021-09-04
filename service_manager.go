package goms

import (
	"context"
	"errors"
	clientV3 "go.etcd.io/etcd/client/v3"
	"goms/utils"
	"log"
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
	ProjectName  = ""
	Services     = make(map[string]*service)
	ServicesLock sync.Mutex
	ServiceId    string
	ServiceName  string
)

func InitProjectName(name string) {
	ProjectName = name
}

func InitServiceId(projectName, rpcPort, httpPort, serviceName string) {
	ServiceName = serviceName
	intranetIp := utils.GetOutboundIP()
	intranetRpcAddr := intranetIp + ":" + rpcPort
	intranetHttpAddr := intranetIp + ":" + httpPort
	ServiceId = projectName + "_" + ServiceName + "_" + intranetRpcAddr + "_" + intranetHttpAddr
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
