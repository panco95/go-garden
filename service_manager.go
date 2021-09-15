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

// Service 服务节点结构体
// PollNext 负载均衡：下一个请求节点索引
// Nodes 服务所有节点
// RequestFinish 请求处理计数
type Service struct {
	PollNext      int
	Nodes         []string
	RequestFinish int
}

// ServiceManager 服务管理操作结构体
type ServiceManager struct {
	Operate     string
	ServiceName string
	ServiceAddr string
}

// ServiceManagerChan 服务管理通道，控制并发安全
var ServiceManagerChan chan ServiceManager

// ServiceId 当前服务ID：唯一性
// ServiceIp 当前服务器地址
// Services 所有服务map
var (
	ServiceId string
	ServiceIp string
	Services  = make(map[string]*Service)
)

// InitService 初始化当前服务
// @param projectName 项目名称
// @param serviceName 服务名称
// @param httpPort http监听端口
// @param rpcPort rpc监听端口
func InitService(projectName, serviceName, httpPort, rpcPort string) error {
	ServiceIp = GetOutboundIP()
	intranetRpcAddr := ServiceIp + ":" + rpcPort
	intranetHttpAddr := ServiceIp + ":" + httpPort
	ServiceId = projectName + "_" + serviceName + "_" + intranetRpcAddr + "_" + intranetHttpAddr

	ServiceManagerChan = make(chan ServiceManager, 0)
	go ServiceManageWatch(ServiceManagerChan)

	return ServiceRegister()
}

// ServiceRegister 注册当前服务
func ServiceRegister() error {
	// 新建租约
	resp, err := etcd.GetClient().Grant(context.TODO(), 2)
	if err != nil {
		return err
	}
	// 授予租约
	if err != nil {
		return err
	}
	_, err = etcd.GetClient().Put(context.TODO(), ServiceId, "0", clientV3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	// keep-alive
	ch, err := etcd.GetClient().KeepAlive(context.TODO(), resp.ID)
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

// ServiceWatcher 服务节点上下线监听
func ServiceWatcher() {
	rch := etcd.GetClient().Watch(context.Background(), Config.ProjectName+"_", clientV3.WithPrefix())
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

// GetAllServices 获取当前所有在线服务节点
// @return []string 服务节点数组
func GetAllServices() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := etcd.GetClient().Get(ctx, Config.ProjectName+"_", clientV3.WithPrefix())
	cancel()
	if err != nil {
		Logger.Debugf("[%s] %s", "GetAllServices", err)
		return []string{}
	}
	var services []string
	for _, ev := range resp.Kvs {
		arr := strings.Split(string(ev.Key), Config.ProjectName+"_")
		service := arr[1]
		services = append(services, service)
	}
	return services
}

// GetServicesByName 通过服务名称获取当前在线服务节点
// @return []string 服务节点数组
func GetServicesByName(serviceName string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := etcd.GetClient().Get(ctx, Config.ProjectName+"_"+serviceName, clientV3.WithPrefix())
	cancel()
	if err != nil {
		Logger.Debugf("[%s] %s", "GetServicesByName", err)
		return []string{}
	}
	var services []string
	for _, ev := range resp.Kvs {
		arr := strings.Split(string(ev.Key), Config.ProjectName+"_"+serviceName+"_")
		serviceAddr := arr[1]
		services = append(services, serviceAddr)
	}
	return services
}

// AddServiceNode 添加服务节点
// @param name 服务名称
// @param addr 服务地址
func AddServiceNode(name, addr string) {
	sm := ServiceManager{
		Operate:     "addNode",
		ServiceName: name,
		ServiceAddr: addr,
	}
	ServiceManagerChan <- sm
}

// DelServiceNode 删除服务节点
// @param name 服务名称
// @param addr 服务地址
func DelServiceNode(name, addr string) {
	sm := ServiceManager{
		Operate:     "delNode",
		ServiceName: name,
		ServiceAddr: addr,
	}
	ServiceManagerChan <- sm
}

// CreateServiceKey 创建服务名称key，避免添加Node报错
// @param 服务名称
func CreateServiceKey(name string) {
	if !ExistsService(name) {
		Services[name] = &Service{
			PollNext:      0,
			Nodes:         []string{},
			RequestFinish: 0,
		}
	}
}

// ExistsService 判断服务名称key是否存在服务管理map中
// @param 服务名称
// @return bool true || false
func ExistsService(name string) bool {
	_, ok := Services[name]
	return ok
}

// GetServiceRpcAddr 获取某个服务节点rpc地址
// @param name 服务名称
// @param index 服务节点数组索引
// @return string 服务节点rpc地址
func GetServiceRpcAddr(name string, index int) (string, error) {
	if index > len(Services[name].Nodes)-1 {
		return "", errors.New("Service not found")
	}
	arr := strings.Split(Services[name].Nodes[index], "_")
	return arr[0], nil
}

// GetServiceHttpAddr 获取某个服务节点http地址
// @param name 服务名称
// @param index 服务节点数组索引
// @return string 服务节点http地址
func GetServiceHttpAddr(name string, index int) (string, error) {
	if index > len(Services[name].Nodes)-1 {
		return "", errors.New("service node not found")
	}
	arr := strings.Split(Services[name].Nodes[index], "_")
	return arr[1], nil
}

// ServiceManageWatch 监听服务管理通道
// @param ch 服务管理通道chan
func ServiceManageWatch(ch chan ServiceManager) {
	for {
		select {
		case sm := <-ch:
			switch sm.Operate {

			case "addNode": // 添加服务节点
				CreateServiceKey(sm.ServiceName)
				Services[sm.ServiceName].Nodes = append(Services[sm.ServiceName].Nodes, sm.ServiceAddr)
				pingServiceRpc(sm.ServiceName)
				break

			case "delNode": // 删除服务节点
				if ExistsService(sm.ServiceName) {
					for i := 0; i < len(Services[sm.ServiceName].Nodes); i++ {
						if Services[sm.ServiceName].Nodes[i] == sm.ServiceAddr {
							Services[sm.ServiceName].Nodes = append(Services[sm.ServiceName].Nodes[:i], Services[sm.ServiceName].Nodes[i+1:]...)
							i--
						}
					}
				}
				break

			case "pullNext": // 服务节点pullNext
				if ExistsService(sm.ServiceName) {
					serviceNum := len(Services[sm.ServiceName].Nodes)
					index := Services[sm.ServiceName].PollNext
					if index >= serviceNum-1 {
						Services[sm.ServiceName].PollNext = 0
					} else {
						Services[sm.ServiceName].PollNext = index + 1
					}
					Services[sm.ServiceName].RequestFinish++
				}
				break
			}
		}
	}
}

// SelectServiceHttpAddr 根据服务名称选择服务节点
// @Description 负载均衡：轮询策略选择服务节点
// @Param name 服务名称
// @return string 服务节点http地址
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

// pingServiceRpc Ping其他服务Rpc地址
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
