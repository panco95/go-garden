package core

import (
	"context"
	"errors"
	"fmt"
	clientV3 "go.etcd.io/etcd/client/v3"
	"math/rand"
	"strings"
	"time"
)

type node struct {
	Addr    string
	Waiting int
	Finish  int64
}

type service struct {
	Nodes []node
}

type serviceOperate struct {
	operate     string
	serviceName string
	serviceAddr string
	nodeIndex   int
}

func (g *Garden) initService(serviceName, httpPort, rpcPort string) error {
	g.Services = map[string]*service{}
	var err error
	g.ServiceIp, err = getOutboundIP()
	if err != nil {
		return err
	}
	g.serviceId = "garden" + "_" + serviceName + "_" + g.ServiceIp + ":" + httpPort + ":" + rpcPort

	g.serviceManager = make(chan serviceOperate, 0)
	go g.RebootFunc("serviceManageWatchReboot", func() {
		g.serviceManageWatch(g.serviceManager)
	})

	return g.serviceRegister()
}

func (g *Garden) serviceRegister() error {
	// New lease
	resp, err := g.etcd.Grant(context.TODO(), 2)
	if err != nil {
		return err
	}
	// The lease was granted
	if err != nil {
		return err
	}
	_, err = g.etcd.Put(context.TODO(), g.serviceId, "0", clientV3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	// keep alive
	ch, err := g.etcd.KeepAlive(context.TODO(), resp.ID)
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
		serviceHttpAddr := arr[1]

		g.addServiceNode(serviceName, serviceHttpAddr)
	}

	go g.RebootFunc("serviceWatcherReboot", g.serviceWatcher)
	go func() {
		for {
			time.Sleep(15 * time.Second)
			g.getAllServices()
		}
	}()

	return nil
}

func (g *Garden) serviceWatcher() {
	rch := g.etcd.Watch(context.Background(), g.cfg.Service.EtcdKey+"_", clientV3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			arr := strings.Split(string(ev.Kv.Key), "_")
			serviceName := arr[1]
			httpAddr := arr[2]
			serviceAddr := httpAddr
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
	resp, err := g.etcd.Get(ctx, g.cfg.Service.EtcdKey+"_", clientV3.WithPrefix())
	cancel()
	if err != nil {
		g.Log(ErrorLevel, "GetAllServices", err)
		return []string{}
	}
	var services []string
	for _, ev := range resp.Kvs {
		arr := strings.Split(string(ev.Key), g.cfg.Service.EtcdKey+"_")
		service := arr[1]
		services = append(services, service)
	}
	return services
}

func (g *Garden) getServicesByName(serviceName string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := g.etcd.Get(ctx, g.cfg.Service.EtcdKey+"_"+serviceName, clientV3.WithPrefix())
	cancel()
	if err != nil {
		g.Log(ErrorLevel, "GetServicesByName", err)
		return []string{}
	}
	var services []string
	for _, ev := range resp.Kvs {
		arr := strings.Split(string(ev.Key), g.cfg.Service.EtcdKey+"_"+serviceName+"_")
		serviceAddr := arr[1]
		services = append(services, serviceAddr)
	}
	return services
}

func (g *Garden) addServiceNode(name, addr string) {
	sm := serviceOperate{
		operate:     "addNode",
		serviceName: name,
		serviceAddr: addr,
	}
	g.serviceManager <- sm
}

func (g *Garden) delServiceNode(name, addr string) {
	sm := serviceOperate{
		operate:     "delNode",
		serviceName: name,
		serviceAddr: addr,
	}
	g.serviceManager <- sm
}

func (g *Garden) createServiceIndex(name string) {
	if !g.existsService(name) {
		g.Services[name] = &service{
			Nodes: []node{},
		}
	}
}

func (g *Garden) existsService(name string) bool {
	_, ok := g.Services[name]
	return ok
}

func (g *Garden) getServiceHttpAddr(name string, index int) (string, error) {
	if index > len(g.Services[name].Nodes)-1 {
		return "", errors.New("service node not found")
	}
	arr := strings.Split(strings.Split(g.Services[name].Nodes[index].Addr, "_")[0], ":")
	return arr[0] + ":" + arr[1], nil
}

func (g *Garden) getServiceRpcAddr(name string, index int) (string, error) {
	if index > len(g.Services[name].Nodes)-1 {
		return "", errors.New("service node not found")
	}
	arr := strings.Split(strings.Split(g.Services[name].Nodes[index].Addr, "_")[0], ":")
	return arr[0] + ":" + arr[2], nil
}

func (g *Garden) serviceManageWatch(ch chan serviceOperate) {
	for {
		select {
		case sm := <-ch:
			switch sm.operate {

			case "addNode":
				g.createServiceIndex(sm.serviceName)
				g.Services[sm.serviceName].Nodes = append(g.Services[sm.serviceName].Nodes, node{Addr: sm.serviceAddr})
				break

			case "delNode":
				if g.existsService(sm.serviceName) {
					for i := 0; i < len(g.Services[sm.serviceName].Nodes); i++ {
						if g.Services[sm.serviceName].Nodes[i].Addr == sm.serviceAddr {
							g.Services[sm.serviceName].Nodes = append(g.Services[sm.serviceName].Nodes[:i], g.Services[sm.serviceName].Nodes[i+1:]...)
							i--
						}
					}
				}
				break

			case "incWaiting":
				if g.existsService(sm.serviceName) {
					g.Services[sm.serviceName].Nodes[sm.nodeIndex].Waiting++
				}
				break

			case "decWaiting":
				if g.existsService(sm.serviceName) {
					g.Services[sm.serviceName].Nodes[sm.nodeIndex].Waiting--
					g.Services[sm.serviceName].Nodes[sm.nodeIndex].Finish++
				}
				break
			}
		}
	}
}

func (g *Garden) selectService(name string) (string, int, error) {
	if _, ok := g.Services[name]; !ok {
		return "", 0, errors.New("service not found")
	}

	waitingMin := 0
	nodeIndex := 0
	nodeLen := len(g.Services[name].Nodes)
	if nodeLen < 1 {
		return "", 0, errors.New("service node not found")
	} else if nodeLen > 1 {
		// get the min waiting service node
		for k, v := range g.Services[name].Nodes {
			if k == 0 {
				waitingMin = v.Waiting
				continue
			}
			if v.Waiting < waitingMin {
				nodeIndex = k
				waitingMin = v.Waiting
			}
		}
		// if all zero, use rand
		if waitingMin == 0 {
			nodeIndex = rand.Intn(nodeLen)
		}
	}

	return g.Services[name].Nodes[nodeIndex].Addr, nodeIndex, nil
}
