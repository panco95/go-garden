package garden

import (
	"context"
	"errors"
	"fmt"
	"garden/drives/etcd"
	"garden/drives/ping"
	clientV3 "go.etcd.io/etcd/client/v3"
	"log"
	"strings"
	"time"
)

type Service struct {
	PollNext int
	Nodes    []string
}

type ServiceManager struct {
	Operate     string
	ServiceName string
	ServiceAddr string
}

var ServiceManagerChan chan ServiceManager

var (
	ServiceId string
	ServiceIp string
	Services  = make(map[string]*Service)
)

func InitService(serviceName, httpPort, rpcPort string) error {
	ServiceIp = GetOutboundIP()
	intranetRpcAddr := ServiceIp + ":" + rpcPort
	intranetHttpAddr := ServiceIp + ":" + httpPort
	ServiceId = "garden" + "_" + serviceName + "_" + intranetRpcAddr + "_" + intranetHttpAddr

	ServiceManagerChan = make(chan ServiceManager, 0)
	go ServiceManageWatch(ServiceManagerChan)

	return ServiceRegister()
}

func ServiceRegister() error {
	// New lease
	resp, err := etcd.Client().Grant(context.TODO(), 2)
	if err != nil {
		return err
	}
	// The lease was granted
	if err != nil {
		return err
	}
	_, err = etcd.Client().Put(context.TODO(), ServiceId, "0", clientV3.WithLease(resp.ID))
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

	services := GetAllServices()
	for _, service := range services {
		arr := strings.Split(service, "_")
		serviceName := arr[0]
		serviceRpcAddr := arr[1]
		serviceHttpAddr := arr[2]

		AddServiceNode(serviceName, serviceRpcAddr+"_"+serviceHttpAddr)
	}

	go ServiceWatcher()

	return nil
}

func ServiceWatcher() {
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
				AddServiceNode(serviceName, serviceAddr)
				log.Printf("[%s] node [%s] join \n", serviceName, serviceAddr)
			case 1: //delete
				DelServiceNode(serviceName, serviceAddr)
				log.Printf("[%s] node [%s] leave \n", serviceName, serviceAddr)
			}
		}
	}
}

func GetAllServices() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := etcd.Client().Get(ctx, "garden_", clientV3.WithPrefix())
	cancel()
	if err != nil {
		Logger.Debugf("[%s] %s", "GetAllServices", err)
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

func GetServicesByName(serviceName string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := etcd.Client().Get(ctx, "garden_"+serviceName, clientV3.WithPrefix())
	cancel()
	if err != nil {
		Logger.Debugf("[%s] %s", "GetServicesByName", err)
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

func AddServiceNode(name, addr string) {
	sm := ServiceManager{
		Operate:     "addNode",
		ServiceName: name,
		ServiceAddr: addr,
	}
	ServiceManagerChan <- sm
}

func DelServiceNode(name, addr string) {
	sm := ServiceManager{
		Operate:     "delNode",
		ServiceName: name,
		ServiceAddr: addr,
	}
	ServiceManagerChan <- sm
}

func CreateServiceIndex(name string) {
	if !ExistsService(name) {
		Services[name] = &Service{
			PollNext: 0,
			Nodes:    []string{},
		}
	}
}

func ExistsService(name string) bool {
	_, ok := Services[name]
	return ok
}

func GetServiceRpcAddr(name string, index int) (string, error) {
	if index > len(Services[name].Nodes)-1 {
		return "", errors.New("Service not found")
	}
	arr := strings.Split(Services[name].Nodes[index], "_")
	return arr[0], nil
}

func GetServiceHttpAddr(name string, index int) (string, error) {
	if index > len(Services[name].Nodes)-1 {
		return "", errors.New("service node not found")
	}
	arr := strings.Split(Services[name].Nodes[index], "_")
	return arr[1], nil
}

func ServiceManageWatch(ch chan ServiceManager) {
	for {
		select {
		case sm := <-ch:
			switch sm.Operate {

			case "addNode":
				CreateServiceIndex(sm.ServiceName)
				Services[sm.ServiceName].Nodes = append(Services[sm.ServiceName].Nodes, sm.ServiceAddr)
				pingServiceRpc(sm.ServiceName)
				break

			case "delNode":
				if ExistsService(sm.ServiceName) {
					for i := 0; i < len(Services[sm.ServiceName].Nodes); i++ {
						if Services[sm.ServiceName].Nodes[i] == sm.ServiceAddr {
							Services[sm.ServiceName].Nodes = append(Services[sm.ServiceName].Nodes[:i], Services[sm.ServiceName].Nodes[i+1:]...)
							i--
						}
					}
				}
				break

			case "pullNext":
				if ExistsService(sm.ServiceName) {
					serviceNum := len(Services[sm.ServiceName].Nodes)
					index := Services[sm.ServiceName].PollNext
					if index >= serviceNum-1 {
						Services[sm.ServiceName].PollNext = 0
					} else {
						Services[sm.ServiceName].PollNext = index + 1
					}
				}
				break
			}
		}
	}
}

func SelectServiceHttpAddr(name string) (string, error) {
	if _, ok := Services[name]; !ok {
		return "", errors.New("service key not found")
	}
	serviceHttpAddr, err := GetServiceHttpAddr(name, Services[name].PollNext)
	if err != nil {
		return "", err
	}

	sm := ServiceManager{
		Operate:     "pullNext",
		ServiceName: Config.ServiceName,
	}
	ServiceManagerChan <- sm

	return serviceHttpAddr, nil
}

func pingServiceRpc(serviceName string) {
	var l string
	rpcAddress, err := GetServiceRpcAddr(serviceName, len(Services[serviceName].Nodes)-1)
	if err != nil {
		l = fmt.Sprintf("[%s][%s %s] %s", "GetSericeRpcAddr", serviceName, rpcAddress, err)
		log.Print(l)
		Logger.Errorf(l)
		return
	}
	if rpcAddress != ServiceIp+":"+Config.RpcPort {
		s, err := ping.Ping(rpcAddress)
		if err != nil {
			l = fmt.Sprintf("[%s][%s %s] %s", "PingRpc", serviceName, rpcAddress, err)
			log.Print(l)
			Logger.Errorf(l)
			return
		}
		l = fmt.Sprintf("[%s][%s %s] %s", "PingRpc", serviceName, rpcAddress, s)
		log.Print(l)
		Logger.Debugf(l)
	}
}
