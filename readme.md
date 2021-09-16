# Go Garden [![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Go Garden是一款面向分布式系统架构的微服务框架

## 概念

Go Garden为分布式系统架构的开发提供了核心需求，包括微服务的基础架构支持，例如gateway网关模块做路由分发支持，服务调用链路追踪的集成。

## 特性

- **服务注册发现**

- **网关路由分发**

- **负载均衡**

- **动态配置**

- **安全认证**

- **服务重试机制**

- **分布式链路追踪**

- **可选组件：Rabbitmq、Redis、Elasticsearch**

- **HTTP服务基于GIN开发，方便广大开发者进行开发**


## 快速开始

`go get -u github.com/panco95/go-garden`

```golang
import "github.com/panco95/go-garden"

// initialise
garden.Init()
// start the service
garden.Run(Route, Auth)
```

访问 [examples](https://github.com/panco95/go-garden/tree/master/examples) 查看详细使用示例

## 基于Go Garden快速构建微服务

#### 1. 准备工作

> Go Garden基于Etcd实现服务注册发现，基于Zipkin实现服务链路追踪，所以需要成功启动必须安装好Etcd和Zipkin

* 在这里给不熟悉的同学介绍Docker快速安装
* 准备好一个Linux系统虚拟机，且安装好Docker


* 启动Etcd：
```
docker run -it -d --name etcd -p 2379:2379 -e "ALLOW_NONE_AUTHENTICATION=yes" -e "ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379" bitnami/etcd
```
* 启动Zipkin：
```
docker run -it -d --name zipkin -p 9411:9411 openzipkin/zipkin
```

#### 2. Gateway网关服务

创建quick_gateway目录后进入目录

执行 `go mod init qucik_gateway` 初始化项目

新建go程序入口文件 `main.go` 并输入以下代码：

```golang
package main

import (
	"github.com/panco95/go-garden"
	"github.com/gin-gonic/gin"
)

func main() {
	// server init
	garden.Init()
	// server run
	garden.Run(garden.GatewayRoute, Auth)
}

// Auth Customize the global middleware
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// before logic
		c.Next()
		// after logic
	}
}
```
安装go mod包： `go mod tidy`

执行程序：`go run main.go`

这时候程序会报一个错误且异常退出，因为没办法继续执行下去：

```
PS D:\code_self\quick_gateway> go run .\main.go
2021/09/15 16:44:32 [Config] Config File "config" Not Found in "[D:\\code_self\\quick_gateway\\configs]"
exit status 1
```

同时Go Garden会保存日志到项目目录的 `runtime`目录下，记得查看哦！言归正传，看日志内容可知道，是配置config找不到，因为还没有创建配置文件，接下来讲解下配置文件。

在项目根目录创建 `configs` 目录并且在目录下创建配置文件 `config.yml` ，把相关配置输入，记得修改相关配置为你的环境噢，192.168.125.184 是我开发环境Linux虚拟机的ip啦~ ：

```yml
ServiceName: gateway
HttpPort: 8080
RpcPort: 8180
CallServiceKey: garden
EtcdAddress:
  - 192.168.125.184:2379
ZipkinAddress: http://192.168.125.184:9411/api/v2/spans

RedisAddress: 192.168.125.184:6379
#ElasticsearchAddress: http://192.168.125.184:9200
#AmqpAddress: amqp://guest:guest@192.168.125.184:5672
```
下面详细说明了每个配置项的作用：

|        配置项         |                                         说明                                         |
| -------------------- | ------------------------------------------------------------------------------------ |
| ServiceName          | 服务名称                                                                              |
| HttpPort             | http监听端口                                                                          |
| RpcPort              | rpc监听端口                                                                           |
| CallServiceKey       | 服务之间调用的密钥，记住请保持每个服务这个配置相同                                         |
| EtcdAddress          | Etcd地址，填写正确的IP加端口，如果是etcd集群的话可以多行填写对应地址                       |
| ZipkinAddress        | zipkin服务的api地址                                                                   |
| RedisAddress         | redis服务的IP加端口                                                                   |
| ElasticsearchAddress | es服务的地址，示例项目没有用到es，可以备注掉，程序启动的时候也不会去执行连接请求             |
| AmqpAddress          | rabbitmq服务的地址，示例项目没有用到rabbitmq，可以备注掉，程序启动的时候也不会去执行连接请求 |

好了，配置文件创建好了，那么现在再来启动一下程序 `go run main.go` 看看吧！

本以为会开开心心的看到程序启动成功，没想到又报了一个错：

```
PS D:\code_self\quick_gateway> go run .\main.go
2021/09/15 17:08:36 [Config] Config File "routes" Not Found in "[D:\\code_self\\quick_gateway\\configs]"
exit status 1
```

这也是一个配置文件找不到的错误，但是呢，不是上面创建的 `config.yml` ，这次是缺少了 `routes.yml` ，这个配置文件是干嘛的呢？想想看，不管是gateway网关调用下游业务服务还是服务A调用服务B，是不知道下游服务的具体请求地址的，可能只知道他这个接口叫做 `login`，可能完整的地址是 `/api/user/login`，也可能是 `/api/v1/user/login`，那么现在要调用服务B的 `login`，就要根据路由配置来获取具体的请求地址。在传统架构里可能直接把地址写在代码里面，万一某一天服务B修改了这个接口的路由，那么得在所有上游服务修改代码更新为正确地址，这是非常蛋疼的架构！
好了，言归正传，现在来创建路由配置 `routes.yml` 吧！

```yml
routes:
  user:
    login: /login
    exists: /exists
  pay:
    order: /order
```
我大胆的猜测，大家看到这个配置内容好像懂了点什么，但是又不是完全懂，那么我来讲解一下吧，我们搭建的微服务示例就基于这么一个路由配置。
首先第一行是固定的，大家不用修改，看第二行 `user` ，表示的是user服务，因为Go Garden是微服务框架嘛，那么一个项目肯定会拆分到很多的服务，例如 用户中心、支付中心、数据中心等等，这里的 `user` 代表的就是用户中心服务，然后它的下面又两行，分别是 `login: /login` 和 `exists: /exists` ，它们表示的是user服务有两个接口，名称分别为 login 和 exists，也就是冒号前面的是接口名称，冒号后面的是服务具体的接口path，如果服务 `login` 接口路由是 `/api/v1/login` ，那么就要在这里要在这里修改它，`user` 服务监听login接口应该是这样子的： `	r.POST("login", func(c *gin.Context) {...})` ，那么依次类推下面的 `exists` 接口和下面的 `pay` 服务的 `order` 接口都是一样的意思，在这里为了简单就没有写多余的接口path，大家在实际开发项目中是可以增加 `v1` 这样的前缀的，以免未来更新接口不兼容老接口时可以增加 `v2` 前缀，总之，这个配置文件就是为了实现 `服务->接口名->接口路由` 这个规则的。
好了，现在把配置文件创建好后，可以再次启动程序 `go run main.go` ：

```
PS D:\code_self\quick_gateway> go run .\main.go
2021/09/15 17:46:29 [gateway] Http listen on port: 8080
2021/09/15 17:46:29 [gateway] Rpc listen on port: 8180
```
终于把gateway网关服务启动成功啦！根据打印信息，可以看到服务监听了 http和rpc两个端口。
现在可以用postman访问网关服务了，地址： `http://127.0.0.1:8080`，会返回404状态码，因为没有带上路由所以网关找不到怎么请求下游服务的路由配置，带上配置好的路由试试：`http://127.0.0.1/api/user/login`，可以发现网关通过路由配置访问下游服务的格式为 `/api/服务名称/接口名称`，api前缀是固定的，访问这个地址不会返回404了，而是返回500状态码且带上了一个json格式数据：
```json
{
    "status": false
}
```
status是一个bool格式，false说明请求出错了，这时候可以打开runtime日志查看错误日志：
```log
2021-09-16T09:30:21.515+0800	ERROR	go-garden@v0.0.0-20210915075049-1d412199ed03/gateway.go:46	[CallService][user/login] service index not found
```
看出是调用下游服务 `user`的`login`接口出错， `service index not found` 是因为并没有启动`user`服务，所以Go Garden并找不到服务地址，所以根本没法请求到下游的`user`服务，那么我们下面继续启动`user`服务。

## 许可证

Go Garden 包含 Apache 2.0 许可证