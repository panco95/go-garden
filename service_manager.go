package garden

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/panco95/go-garden/drives/etcd"
	"github.com/panco95/go-garden/sync"
	"github.com/panco95/go-garden/utils"
	clientV3 "go.etcd.io/etcd/client/v3"
	"strings"
	"time"
)

type service struct {
	PollNext int
	Nodes    []string
}

type serviceManager struct {
	Operate     string
	ServiceName string
	ServiceAddr string
}

var serviceManagerChan chan serviceManager

var (
	serviceId string
	serviceIp string
	services  = make(map[string]*service)
)

func initService(serviceName, httpPort, rpcPort string) error {
	var err error
	serviceIp, err = utils.GetOutboundIP()
	if err != nil {
		return err
	}
	intranetRpcAddr := serviceIp + ":" + rpcPort
	intranetHttpAddr := serviceIp + ":" + httpPort
	serviceId = "garden" + "_" + serviceName + "_" + intranetRpcAddr + "_" + intranetHttpAddr

	serviceManagerChan = make(chan serviceManager, 0)
	go serviceManageWatch(serviceManagerChan)

	return serviceRegister()
}

func serviceRegister() error {
	// New lease
	resp, err := etcd.Client().Grant(context.TODO(), 2)
	if err != nil {
		return err
	}
	// The lease was granted
	if err != nil {
		return err
	}
	_, err = etcd.Client().Put(context.TODO(), serviceId, "0", clientV3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	// keep alive
	ch, err := etcd.Client().KeepAlive(context.TODO(), resp.ID)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ch:

			}
		}
	}()

	services := getAllServices()
	for _, service := range services {
		arr := strings.Split(service, "_")
		serviceName := arr[0]
		serviceRpcAddr := arr[1]
		serviceHttpAddr := arr[2]

		addServiceNode(serviceName, serviceRpcAddr+"_"+serviceHttpAddr)
	}

	go serviceWatcher()

	return nil
}

func serviceWatcher() {
	rch := etcd.Client().Watch(context.Background(), "garden_", clientV3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			arr := strings.Split(string(ev.Kv.Key), "_")
			serviceName := arr[1]
			rpcAddr := arr[2]
			httpAddr := arr[3]
			serviceAddr := rpcAddr + "_" + httpAddr
			switch ev.Type {
			case 0: //put
				addServiceNode(serviceName, serviceAddr)
				Log(InfoLevel, "Service", fmt.Sprintf("[%s] node [%s] join", serviceName, serviceAddr))
			case 1: //delete
				delServiceNode(serviceName, serviceAddr)
				Log(InfoLevel, "Service", fmt.Sprintf("[%s] node [%s] leave", serviceName, serviceAddr))
			}
		}
	}
}

func getAllServices() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := etcd.Client().Get(ctx, "garden_", clientV3.WithPrefix())
	cancel()
	if err != nil {
		Log(ErrorLevel, "GetAllServices", err)
		return []string{}
	}
	var services []string
	for _, ev := range resp.Kvs {
		arr := strings.Split(string(ev.Key), "garden_")
		service := arr[1]
		services = append(services, service)
	}
	return services
}

func getServicesByName(serviceName string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := etcd.Client().Get(ctx, "garden_"+serviceName, clientV3.WithPrefix())
	cancel()
	if err != nil {
		Log(ErrorLevel, "GetServicesByName", err)
		return []string{}
	}
	var services []string
	for _, ev := range resp.Kvs {
		arr := strings.Split(string(ev.Key), "garden_"+serviceName+"_")
		serviceAddr := arr[1]
		services = append(services, serviceAddr)
	}
	return services
}

func addServiceNode(name, addr string) {
	sm := serviceManager{
		Operate:     "addNode",
		ServiceName: name,
		ServiceAddr: addr,
	}
	serviceManagerChan <- sm
}

func delServiceNode(name, addr string) {
	sm := serviceManager{
		Operate:     "delNode",
		ServiceName: name,
		ServiceAddr: addr,
	}
	serviceManagerChan <- sm
}

func createServiceIndex(name string) {
	if !existsService(name) {
		services[name] = &service{
			PollNext: 0,
			Nodes:    []string{},
		}
	}
}

func existsService(name string) bool {
	_, ok := services[name]
	return ok
}

func getServiceRpcAddr(name string, index int) (string, error) {
	if index > len(services[name].Nodes)-1 {
		return "", errors.New("Service not found")
	}
	arr := strings.Split(services[name].Nodes[index], "_")
	return arr[0], nil
}

func getServiceHttpAddr(name string, index int) (string, error) {
	if index > len(services[name].Nodes)-1 {
		return "", errors.New("service node not found")
	}
	arr := strings.Split(services[name].Nodes[index], "_")
	return arr[1], nil
}

func serviceManageWatch(ch chan serviceManager) {
	for {
		select {
		case sm := <-ch:
			switch sm.Operate {

			case "addNode":
				createServiceIndex(sm.ServiceName)
				services[sm.ServiceName].Nodes = append(services[sm.ServiceName].Nodes, sm.ServiceAddr)
				break

			case "delNode":
				if existsService(sm.ServiceName) {
					for i := 0; i < len(services[sm.ServiceName].Nodes); i++ {
						if services[sm.ServiceName].Nodes[i] == sm.ServiceAddr {
							services[sm.ServiceName].Nodes = append(services[sm.ServiceName].Nodes[:i], services[sm.ServiceName].Nodes[i+1:]...)
							i--
						}
					}
				}
				break

			case "pullNext":
				if existsService(sm.ServiceName) {
					serviceNum := len(services[sm.ServiceName].Nodes)
					index := services[sm.ServiceName].PollNext
					if index >= serviceNum-1 {
						services[sm.ServiceName].PollNext = 0
					} else {
						services[sm.ServiceName].PollNext = index + 1
					}
				}
				break
			}
		}
	}
}

func selectServiceHttpAddr(name string) (string, error) {
	if _, ok := services[name]; !ok {
		return "", errors.New("service index not found")
	}
	serviceHttpAddr, err := getServiceHttpAddr(name, services[name].PollNext)
	if err != nil {
		return "", err
	}

	sm := serviceManager{
		Operate:     "pullNext",
		ServiceName: name,
	}
	serviceManagerChan <- sm

	return serviceHttpAddr, nil
}

// SyncRoutes sync routes.yml to other each service
func syncRoutes() {
	fileData, err := utils.ReadFile("configs/routes.yml")
	if err != nil {
		Log(ErrorLevel, "SyncRoutes", err)
		return
	}

	if bytes.Compare(syncCache, fileData) == 0 {
		return
	}

	syncCache = fileData

	for k, v := range services {
		for i := 0; i < len(v.Nodes); i++ {
			serviceRpcAddress, err := getServiceRpcAddr(k, i)
			if err != nil {
				Log(ErrorLevel, "GetServiceRpcAddr", err)
				continue
			}
			if strings.Compare(serviceRpcAddress, fmt.Sprintf("%s:%s", serviceIp, Config.RpcPort)) != 0 {
				result, err := sync.SendSyncRoutes(serviceRpcAddress, fileData)
				if err != nil {
					Log(ErrorLevel, "SendSyncRoutes", err)
					continue
				}
				if result != true {
					Log(ErrorLevel, "SendSyncRoutesResult", "false")
				}
				Log(InfoLevel, "SendSyncRoutesResult", "true")
			}
		}
	}
}
