package core

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

type serviceOperate struct {
	Operate     string
	ServiceName string
	ServiceAddr string
}

func (g *Garden) initService(serviceName, httpPort, rpcPort string) error {
	g.services = map[string]*service{}
	var err error
	g.serviceIp, err = utils.GetOutboundIP()
	if err != nil {
		return err
	}
	intranetRpcAddr := g.serviceIp + ":" + rpcPort
	intranetHttpAddr := g.serviceIp + ":" + httpPort
	g.serviceId = "garden" + "_" + serviceName + "_" + intranetRpcAddr + "_" + intranetHttpAddr

	g.serviceManager = make(chan serviceOperate, 0)
	go g.serviceManageWatch(g.serviceManager)

	return g.serviceRegister()
}

func (g *Garden) serviceRegister() error {
	// New lease
	resp, err := etcd.Client().Grant(context.TODO(), 2)
	if err != nil {
		return err
	}
	// The lease was granted
	if err != nil {
		return err
	}
	_, err = etcd.Client().Put(context.TODO(), g.serviceId, "0", clientV3.WithLease(resp.ID))
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

	services := g.getAllServices()
	for _, service := range services {
		arr := strings.Split(service, "_")
		serviceName := arr[0]
		serviceRpcAddr := arr[1]
		serviceHttpAddr := arr[2]

		g.addServiceNode(serviceName, serviceRpcAddr+"_"+serviceHttpAddr)
	}

	go g.serviceWatcher()

	return nil
}

func (g *Garden) serviceWatcher() {
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
				g.addServiceNode(serviceName, serviceAddr)
				g.Log(InfoLevel, "Service", fmt.Sprintf("[%s] node [%s] join", serviceName, serviceAddr))
			case 1: //delete
				g.delServiceNode(serviceName, serviceAddr)
				g.Log(InfoLevel, "Service", fmt.Sprintf("[%s] node [%s] leave", serviceName, serviceAddr))
			}
		}
	}
}

func (g *Garden) getAllServices() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := etcd.Client().Get(ctx, "garden_", clientV3.WithPrefix())
	cancel()
	if err != nil {
		g.Log(ErrorLevel, "GetAllServices", err)
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

func (g *Garden) getServicesByName(serviceName string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := etcd.Client().Get(ctx, "garden_"+serviceName, clientV3.WithPrefix())
	cancel()
	if err != nil {
		g.Log(ErrorLevel, "GetServicesByName", err)
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

func (g *Garden) addServiceNode(name, addr string) {
	sm := serviceOperate{
		Operate:     "addNode",
		ServiceName: name,
		ServiceAddr: addr,
	}
	g.serviceManager <- sm
}

func (g *Garden) delServiceNode(name, addr string) {
	sm := serviceOperate{
		Operate:     "delNode",
		ServiceName: name,
		ServiceAddr: addr,
	}
	g.serviceManager <- sm
}

func (g *Garden) createServiceIndex(name string) {
	if !g.existsService(name) {
		g.services[name] = &service{
			PollNext: 0,
			Nodes:    []string{},
		}
	}
}

func (g *Garden) existsService(name string) bool {
	_, ok := g.services[name]
	return ok
}

func (g *Garden) getServiceRpcAddr(name string, index int) (string, error) {
	if index > len(g.services[name].Nodes)-1 {
		return "", errors.New("Service not found")
	}
	arr := strings.Split(g.services[name].Nodes[index], "_")
	return arr[0], nil
}

func (g *Garden) getServiceHttpAddr(name string, index int) (string, error) {
	if index > len(g.services[name].Nodes)-1 {
		return "", errors.New("service node not found")
	}
	arr := strings.Split(g.services[name].Nodes[index], "_")
	return arr[1], nil
}

func (g *Garden) serviceManageWatch(ch chan serviceOperate) {
	for {
		select {
		case sm := <-ch:
			switch sm.Operate {

			case "addNode":
				g.createServiceIndex(sm.ServiceName)
				g.services[sm.ServiceName].Nodes = append(g.services[sm.ServiceName].Nodes, sm.ServiceAddr)
				break

			case "delNode":
				if g.existsService(sm.ServiceName) {
					for i := 0; i < len(g.services[sm.ServiceName].Nodes); i++ {
						if g.services[sm.ServiceName].Nodes[i] == sm.ServiceAddr {
							g.services[sm.ServiceName].Nodes = append(g.services[sm.ServiceName].Nodes[:i], g.services[sm.ServiceName].Nodes[i+1:]...)
							i--
						}
					}
				}
				break

			case "pullNext":
				if g.existsService(sm.ServiceName) {
					serviceNum := len(g.services[sm.ServiceName].Nodes)
					index := g.services[sm.ServiceName].PollNext
					if index >= serviceNum-1 {
						g.services[sm.ServiceName].PollNext = 0
					} else {
						g.services[sm.ServiceName].PollNext = index + 1
					}
				}
				break
			}
		}
	}
}

func (g *Garden) selectServiceHttpAddr(name string) (string, error) {
	if _, ok := g.services[name]; !ok {
		return "", errors.New("service index not found")
	}
	serviceHttpAddr, err := g.getServiceHttpAddr(name, g.services[name].PollNext)
	if err != nil {
		return "", err
	}

	sm := serviceOperate{
		Operate:     "pullNext",
		ServiceName: name,
	}
	g.serviceManager <- sm

	return serviceHttpAddr, nil
}

// SyncRoutes sync routes.yml to other each service
func (g *Garden) syncRoutes() {
	fileData, err := utils.ReadFile("configs/routes.yml")
	if err != nil {
		g.Log(ErrorLevel, "SyncRoutes", err)
		return
	}

	if bytes.Compare(g.syncCache, fileData) == 0 {
		return
	}

	g.syncCache = fileData

	for k, v := range g.services {
		for i := 0; i < len(v.Nodes); i++ {
			serviceRpcAddress, err := g.getServiceRpcAddr(k, i)
			if err != nil {
				g.Log(ErrorLevel, "GetServiceRpcAddr", err)
				continue
			}
			if strings.Compare(serviceRpcAddress, fmt.Sprintf("%s:%s", g.serviceIp, g.Cfg.RpcPort)) != 0 {
				result, err := sync.SendSyncRoutes(serviceRpcAddress, fileData)
				if err != nil {
					g.Log(ErrorLevel, "SendSyncRoutes", err)
					continue
				}
				if result != true {
					g.Log(ErrorLevel, "SendSyncRoutesResult", "false")
				}
				g.Log(InfoLevel, "SendSyncRoutesResult", "true")
			}
		}
	}
}