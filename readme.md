**GO-MS：GO实现的微服务基础框架**
<br>
<br>
说明：本项目刚启动，目标是实现微服务的基本架构，当前着力于开发gateway入口，也就是api网关。<br>

**准备工作：**<br>
1、安装etcd，Docker快捷安装：<br>
`docker run --rm -it -d -p 2379:2379 --env ALLOW_NONE_AUTHENTICATION=yes --env ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379 bitnami/etcd`
<br>

**文档**<br><br>
**一、Gateway**<br>

启动：`go run cmd/gateway/main.go`<br>
参数：<br>
-http_port：http监听端口<br>
-rpc_port：rpc监听端口<br>
-etcd_addr：etcd服务地址（支持集群格式：`127.0.0.1:2379|127.0.0.1:2380`）<br>
-version：打印版本<br>

模拟三个Gateway网关节点进行集群架设：<br>
`1、go run cmd/gateway/main.go -http_port 8080 -rpc_port 8180 -etcd_addr 127.0.0.1:2379`<br>
`2、go run cmd/gateway/main.go -http_port 8081 -rpc_port 8181 -etcd_addr 127.0.0.1:2379`<br>
`3、go run cmd/gateway/main.go -http_port 8082 -rpc_port 8182 -etcd_addr 127.0.0.1:2379`<br>

等待几秒钟，观察打印如下：<br>
`Server [gateway] cluster [192.168.125.179:8182] join`<br>
`Server [gateway] cluster [192.168.125.179:8183] join`<br>
表示gateway节点互相能够发现其他节点服务

<br>

**二、Service**<br>

启动服务节点：<br>
`go run cmd/services/user/main.go -rpc_port 9010 -etcd_addr 127.0.0.1:2379`<br>
参数：<br>
-rpc_port：rpc监听端口<br>
-etcd_addr：etcd服务地址（支持集群格式：`127.0.0.1:2379|127.0.0.1:2380`）<br>
-version：打印版本<br>

模拟三个user服务节点进行集群架设：<br>
`1、 go run cmd/services/user/main.go -rpc_port 9010 -etcd_addr 127.0.0.1:2379`<br>
`2、 go run cmd/services/user/main.go -rpc_port 9011 -etcd_addr 127.0.0.1:2379`<br>
`3、 go run cmd/services/user/main.go -rpc_port 9012 -etcd_addr 127.0.0.1:2379`<br>

等待几秒钟，切换gateway、user任意一个节点的命令行打印信息：<br>
`Server [user] cluster [192.168.125.179:8080] join`<br>
`Server [user] cluster [192.168.125.179:8081] join`<br>
`Server [user] cluster [192.168.125.179:8082] join`<br>

另外我们再看一下第一台gateway打印的所有信息：<br>
`2021/08/09 14:50:36 [Http][gateway service] Listen on port: 8080`<br>
`2021/08/09 14:50:42 Server [gateway] cluster [192.168.125.179:8181] join`<br>
`2021/08/09 14:50:49 Server [gateway] cluster [192.168.125.179:8182] join`<br>
`2021/08/09 14:51:02 Server [user] cluster [192.168.125.179:9010] join`<br>
`2021/08/09 14:51:11 Server [user] cluster [192.168.125.179:9011] join`<br>
`2021/08/09 14:51:18 Server [user] cluster [192.168.125.179:9012] join `<br>

**我们可以总结一下：任意服务的任意节点互相能够发现其他任意服务的任意节点**<br>
**这就是基于etcd实现的服务注册发现模型：**<br>
我们可以去看看etcd存储的数据：<br>
`etcdctl get --prefix go-ms`<br>
（小提示：Docker启动的etcd需要进入etcd容器查看）<br>
`docker container exec -ti [容器id] bash`<br>
打印如下：<br>
`go-ms_gateway_192.168.125.179:8180`<br>
`0`<br>
`go-ms_gateway_192.168.125.179:8181`<br>
`0`<br>
`go-ms_gateway_192.168.125.179:8182`<br>
`0`<br>
`go-ms_user_192.168.125.179:9010`<br>
`0`<br>
`go-ms_user_192.168.125.179:9011`<br>
`0`<br>
`go-ms_user_192.168.125.179:9012`<br>
`0`<br>
可以看出，我们刚刚启动的所以服务节点都在etcd里面以go-ms_前缀存储<br>


