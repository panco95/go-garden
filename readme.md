**GO-MS：GO实现的微服务基础框架**
<br>
<br>
说明：本项目刚启动，目标是实现微服务的基本架构，当前着力于开发gateway入口，也就是api网关。<br>

**准备工作：**<br>
1、安装etcd，Docker快捷安装：<br>
`docker run --rm -it -d -p 2379:2379 --env ALLOW_NONE_AUTHENTICATION=yes --env ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379 bitnami/etcd`
<br>

**文档**<br><br>
**一、api网关（Gateway）**<br>

启动：`go run cmd/gateway/main.go`<br>
参数：<br>
-http_port：http监听端口<br>
-rpc_port：rpc监听端口<br>
-etcd_addr：etcd服务地址（支持集群格式：`127.0.0.1:2379|127.0.0.1:2380`）<br>
-version：打印版本<br>

测试：打开三个命令行模拟三个节点测试gateway集群服务注册发现<br>
`1、go run cmd/gateway/main.go -http_port 8080 -rpc_port 8180 -etcd_addr 127.0.0.1:2379`<br>
`2、go run cmd/gateway/main.go -http_port 8081 -rpc_port 8181 -etcd_addr 127.0.0.1:2379`<br>
`3、go run cmd/gateway/main.go -http_port 8082 -rpc_port 8182 -etcd_addr 127.0.0.1:2379`<br>

等待几秒钟，观察打印如下：<br>
`Server [gateway] cluster [192.168.125.179:8182] join`<br>
`Server [gateway] cluster [192.168.125.179:8183] join`<br>
表示节点已经发现其他节点服务启动

<br>

**二、服务示例（Service）**<br>

启动服务节点：<br>
`go run cmd/services/user/main.go -rpc_port 9010 -etcd_addr 127.0.0.1:2379`<br>
参数：<br>
-rpc_port：rpc监听端口<br>
-etcd_addr：etcd服务地址（支持集群格式：`127.0.0.1:2379|127.0.0.1:2380`）<br>
-version：打印版本<br>