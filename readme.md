**GO-MS：GO微服务基本模板**
<br>
<br>
**说明：**<br>
本项目刚启动，目标是实现微服务的基本模板。<br><br>
**已实现特性：**<br>
1、服务注册发现<br>
2、由网关控制到下游服务集群的负载均衡<br>

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

观察我们第一个gateway节点打印如下：<br>
`2021/08/10 11:08:01 [gateway] Http Listen on port: 8080`<br>
`2021/08/10 11:08:14 [gateway] node [192.168.125.179:8181_192.168.125.179:8081] join`<br>
`2021/08/10 11:08:23 [gateway] node [192.168.125.179:8182_192.168.125.179:8082] join`

表示gateway节点互相能够发现其他节点服务启动

<br>

**二、Service**<br>

启动服务节点：<br>
`go run cmd/services/user/main.go`<br>
参数：<br>
-rpc_port：rpc监听端口<br>
-etcd_addr：etcd服务地址（支持集群格式：`127.0.0.1:2379|127.0.0.1:2380`）<br>
-version：打印版本<br>

模拟三个user服务节点进行集群架设：<br>
`1、 go run cmd/services/user/main.go -http_port 9080 -rpc_port 9180 -etcd_addr 127.0.0.1:2379`<br>
`2、 go run cmd/services/user/main.go -http_port 9081 -rpc_port 9181 -etcd_addr 127.0.0.1:2379`<br>
`3、 go run cmd/services/user/main.go -http_port 9082 -rpc_port 9182 -etcd_addr 127.0.0.1:2379`<br>

观察我们第一个node节点打印如下：<br>
`2021/08/10 11:10:46 [user] Http Listen on port: 9080`<br>
`2021/08/10 11:10:54 [user] node [192.168.125.179:9181_192.168.125.179:9081] join`<br>
`2021/08/10 11:11:02 [user] node [192.168.125.179:9182_192.168.125.179:9082] join`<br>

另外我们再看一下第一个gateway打印的所有信息：<br>
```
2021/08/10 11:08:01 [gateway] Http Listen on port: 8080
2021/08/10 11:08:14 [gateway] node [192.168.125.179:8181_192.168.125.179:8081] join
2021/08/10 11:08:23 [gateway] node [192.168.125.179:8182_192.168.125.179:8082] join
2021/08/10 11:10:46 [user] node [192.168.125.179:9180_192.168.125.179:9080] join
2021/08/10 11:10:54 [user] node [192.168.125.179:9181_192.168.125.179:9081] join
2021/08/10 11:11:02 [user] node [192.168.125.179:9182_192.168.125.179:9082] join
```

现在我们随意关闭一个节点，查看其他所有节点打印信息：<br>
`2021/08/10 11:13:23 [user] node [192.168.125.179:9180_192.168.125.179:9080] leave`<br>
每个服务节点不仅可以监听新节点的加入(join)，还能监听节点的离开(leave)

**我们可以总结一下：任意服务的任意节点互相能够发现其他任意服务的任意节点**<br>
**这就是基于etcd实现的服务注册发现模型：**<br>
我们可以去看看etcd存储的数据：<br>
`etcdctl get --prefix go-ms`<br>
（小提示：Docker启动的etcd需要进入etcd容器查看）<br>
`docker container exec -ti [容器id] bash`<br>
打印如下：<br>
```
go-ms_gateway_192.168.125.179:8180_192.168.125.179:8080
0
go-ms_gateway_192.168.125.179:8181_192.168.125.179:8081
0
go-ms_gateway_192.168.125.179:8182_192.168.125.179:8082
0
go-ms_user_192.168.125.179:9181_192.168.125.179:9081
0
go-ms_user_192.168.125.179:9182_192.168.125.179:9082
0
```
<br>
可以看出，我们刚刚启动的所以服务节点都在etcd里面以go-ms_前缀存储，后面跟着服务名称，然后是两个服务地址，前面的是RPC监听地址，后面的是HTTP监听地址<br><br>

**三、访问API网关**<br>

提示：为了规范接口，api接口全部统一为POST请求，参数格式为 application/json。<br>

1、集群服务状态查看：<br>
打开浏览器，输入gateway任一节点http地址，加上/cluster路由：
`http://127.0.0.1:8082/cluster` <br>

```
{
  "servers": {
    "gateway": {
      "PollNext": 0,
      "Nodes": [
        "192.168.125.179:8180_192.168.125.179:8080",
        "192.168.125.179:8181_192.168.125.179:8081",
        "192.168.125.179:8182_192.168.125.179:8082"
      ],
      "RequestFinish": 0
    },
    "user": {
      "PollNext": 1,
      "Nodes": [
        "192.168.125.179:9180_192.168.125.179:9080",
        "192.168.125.179:9181_192.168.125.179:9081",
        "192.168.125.179:9182_192.168.125.179:9082"
      ],
      "RequestFinish": 0
    }
  }
}
```

可以看到当前整个微服务的所有服务所有节点的信息，PollNext是负载均衡轮询策略索引，Nodes就是当前在线服务集群列表，RequestFinish是当前服务集群一共处理的请求数。<br>

2、打开postman，选择POST请求方式，URL输入gateway任一节点http地址，然后加上访问用户服务的路由如：``
http://127.0.0.1:8082/api/user/login
``，参数使用
``application/json``
，加上json格式数据如：
```
{
"field_a": "hello",
"field_b": "okay",
"field_c": "thanks"
}
```
<br>
正常返回json数据：<br>

```
{
    "code": 200,
    "data": {
        "field_a": "hello",
        "field_b": "okay",
        "field_c": "thanks"
    },
    "message": "success"
}
```

3、再看集群服务状态

```
{
  "servers": {
    "gateway": {
      "PollNext": 0,
      "Nodes": [
        "192.168.125.179:8180_192.168.125.179:8080",
        "192.168.125.179:8181_192.168.125.179:8081",
        "192.168.125.179:8182_192.168.125.179:8082"
      ],
      "RequestFinish": 0
    },
    "user": {
      "PollNext": 1,
      "Nodes": [
        "192.168.125.179:9180_192.168.125.179:9080",
        "192.168.125.179:9181_192.168.125.179:9081",
        "192.168.125.179:9182_192.168.125.179:9082"
      ],
      "RequestFinish": 1
    }
  }
}
```

这是可以看到user集群的的PollNext已经变成1，表示下一个请求将由第二个Node来处理请求。<br>