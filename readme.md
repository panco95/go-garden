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

创建gateway目录后进入目录

执行 `go mod init gateway` 初始化项目

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
PS D:\code_self\gateway> go run .\main.go
2021/09/15 16:44:32 [Config] Config File "config" Not Found in "[D:\\code_self\\gateway\\configs]"
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

#RedisAddress: 192.168.125.184:6379
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
| RedisAddress         | redis服务的IP加端口,示例项目有用到，现在可以暂时备注掉                                    |
| ElasticsearchAddress | es服务的地址，示例项目没有用到es，可以备注掉，程序启动的时候也不会去执行连接请求             |
| AmqpAddress          | rabbitmq服务的地址，示例项目没有用到rabbitmq，可以备注掉，程序启动的时候也不会去执行连接请求 |

好了，配置文件创建好了，那么现在再来启动一下程序 `go run main.go` 看看吧！

本以为会开开心心的看到程序启动成功，没想到又报了一个错：

```
PS D:\code_self\gateway> go run .\main.go
2021/09/15 17:08:36 [Config] Config File "routes" Not Found in "[D:\\code_self\\gateway\\configs]"
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
PS D:\code_self\gateway> go run .\main.go
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

#### 3. User服务
跟gateway服务步骤一样创建好项目`user`和配置文件`config.yml`、`routes.yml`，我们要稍稍改动以下`config.yml`的`ServiceName`，改为`user`，然后创建main.go程序启动入口文件：
```golang
package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden"
	"net/http"
)

func main() {
	garden.Init()
	garden.Run(Route, nil)
}

func Route(r *gin.Engine) {
	r.Use(garden.CheckCallSafeMiddleware()) // 调用接口安全验证
	r.POST("login", Login)
	r.POST("exists", Exists)
}

func Login(c *gin.Context) {
    ...
}
func Exists(c *gin.Context) {
    ...
}

```
观察代码，启动服务的时候有两个参数跟gateway不一样：

1、`garden.Run()`第一个参数是路由，因为gateway的路由是在Go Garden内部集成的，所以`gateway`服务直接使用了`arden.GatewayRoute`，`user`服务需要自己实现路由，就是下面的`Route`函数，这是基于`Gin`框架的路由，第一行`r.Use(garden.CheckCallSafeMiddleware())`这是校验服务调用密钥的中间件，防止客户端跳过`gateway`直接请求`user`，这样的话`gateway`就发挥不了作用了，下面两行`r.Post("login",Login)`和`r.Post("exists",Exists)`就是具体的路由实现，再看看`routes.yml`可以看出是对应上的，假设这么写路由`r.Post("v1/login",Login)`，那在`routes.yml`应该写成`login: /v1/login`;

2、第二个参数是全局中间件，在`gateway`网关服务中需要实现全局鉴权，所以我们添加了一个`Auth`中间件，我们假设`user`不需要单独的鉴权，所里这里直接写`nil`。



这里省略了`Login`和`Exists`的具体逻辑，请查看示例代码参考[examples/user](https://github.com/panco95/go-garden/tree/master/examples/user)

> 注意：示例代码逻辑实现用到了redis，所以我们需要启动redis服务以及在`config.yml`中填写连接地址。

Docker启动redis：

`docker run --rm -it -d --name redis -p 6379:6379 redis`

修改config.yml：
```yml
...
RedisAddress: 192.168.125.184:6379
...
```

一切准备就绪，启动`user`服务：`go run main.go`，查看输出：
```
PS D:\go-garden\examples\user> go run .\main.go
2021/09/16 10:23:44 [PingRpc][gateway 192.168.8.98:8180] ok
2021/09/16 10:23:44 [user] Http listen on port: 8081
2021/09/16 10:23:44 [user] Rpc listen on port: 8181
```
跟`gateway`一样，监听了Http、Rpc两个端口，同时启动服务的时候还做了一件事情，就是发现了上面启动的的`gateway`服务，而且ping了一下`gateway`的Rpc端口保证可正常通信；

接着我们切换到`gateway`服务的窗口，也输出了两行打印信息：

```
2021/09/16 10:23:44 [user] node [192.168.8.98:8181_192.168.8.98:8081] join
2021/09/16 10:23:44 [PingRpc][user 192.168.8.98:8181] ok
```

> 第一行表示`gateway`发现了`user`的一个服务节点；
> 第二行表示`gateway`也ping了一下`user`服务的Rpc端口。

总结一下，这就是Go Garden的`服务自动注册发现`特性，不论你启动多少个服务多少个节点，它们都能互相发现和通信。

现在`user`服务启动成功，现在可以再次使用postman访问`gateway`的`user`服务路由：`http://127.0.0.1:8080/api/user/login`，增加一个请求参数`username`，发送请求，响应如下：

```json
{
    "code": 0,
    "data": null,
    "msg": "登录成功",
    "status": true
}
```
返回参数`status`为`true`表示请求成功，注意，只有`status`参数是`gateway`返回的，以告诉客户端请求是否成功，其他参数都是`user`返回的，这样使得`gateway`和其他服务耦合度降低，数据格式可以右开发者自行设计。

#### 4. 服务集群

Go Garden基于`服务自动注册发现`特性，支持大规模的服务集群，例如`user`服务我们可以启动多个示例，现在我们复制一份`user`服务的代码，修改`config.yml`的两个监听端口防止端口冲突，改好端口后启动第二个`user`服务节点：`go run main.go`，查看输出：

```
2021/09/16 11:00:22 [PingRpc][gateway 192.168.8.98:8180] ok
2021/09/16 11:00:22 [PingRpc][user 192.168.8.98:8181] ok
2021/09/16 11:00:22 [user] Http listen on port: 8280
2021/09/16 11:00:22 [user] Rpc listen on port: 8281
```

这是启动的第二个`user`服务节点，它发现了以及ping了`gateway`节点和第一个`user`节点，切换到`gateway`和`user`节点1，都会输出：

```
2021/09/16 11:00:22 [user] node [192.168.8.98:8281_192.168.8.98:8280] join
2021/09/16 11:00:22 [PingRpc][user 192.168.8.98:8281] ok
```

现在`user`服务右两个节点，我么可以称之为`user`服务集群，那么`gateway`调用`user`服务的时候会是什么一个情况呢？

我们再次使用Postman给`gateway`发送两次请求：`http://127.0.0.1:8080/api/user/login` ，会发现`user`服务节点1和节点2都会打印一次请求日志，这是`gateway`服务控制的下游服务集群的负载均衡；如果是`gateway`服务集群呢，上游可能并没有服务，那么我们建议是增加一层较稳定的`webserver`例如`nginx`，在`nginx`层增加对`gateway`网关的负载均衡。

## 许可证

Go Garden 包含 Apache 2.0 许可证