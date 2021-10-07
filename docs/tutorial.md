## 基于go-garden快速构建微服务

> go-garden的Http服务基于Gin开发，在教程中会涉及到Gin框架的一些内容，例如请求上下文、中间件等，如开发者不了解Gin，请先阅读Gin相关文档！

### 1. 环境准备

go-garden基于Etcd实现服务注册发现，基于Zipkin实现链路追踪，基于消息队列实现自动路由同步，所以需要成功启动必须安装好Etcd、Zipkin、Rabbitmq

* 在这里给不熟悉的同学介绍Docker快速安装
* 准备好一个Linux系统虚拟机，且安装好Docker
* Docker示例环境仅作为测试使用
 
```
docker run -it -d --name etcd -p 2379:2379 -e "ALLOW_NONE_AUTHENTICATION=yes" -e "ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379" bitnami/etcd
docker run -it -d --name zipkin -p 9411:9411 openzipkin/zipkin
docker run -it -d --name rabbitmq --hostname rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

### 2. Gateway Api网关

创建gateway目录后进入目录

执行 `go mod init gateway` 初始化项目

新建go程序入口文件 `main.go` 并输入以下代码：

```golang
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
)

var service *core.Garden

func main() {
	service = core.New()
	service.Run(service.GatewayRoute, Auth)
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

增加go.mod包引用go-garden

```
require (
	github.com/panco95/go-garden v1.0.16
)
```

安装go mod包： `go mod tidy`

执行程序：`go run main.go`

这时候程序会报一个错误且异常退出，因为没办法继续执行下去：

```
PS D:\go-garden-demo\gateway> go run .\main.go
2021/09/28 11:31:22 [Config] Config File "config" Not Found in "[D:\\go-garden-demo\\gateway\\configs]"
exit status 1
```

观察日志内容可知道，找不到config配置文件；

在项目根目录创建 `configs` 目录并且在目录下创建配置文件 `config.yml` ，把相关配置输入，记得修改相关配置为你的环境，192.168.125.185 是我开发环境Linux虚拟机的ip ：

```yml
service:
  debug: true
  serviceName: gateway
  listenOut: 1
  listenPort: 8080
  callKey: garden
  callRetry: 100/200/300
  etcdAddress:
    - 192.168.125.185:2379
  zipkinAddress: http://192.168.125.185:9411/api/v2/spans
  amqpAddress: amqp://guest:guest@192.168.125.185:5672

config:
```

`service`表示go-garden启动必填配置项，`config`表示自定义业务配置项；示例Demo的网关没有全局验证中间件，无需增加业务配置。

下面详细说明了service每个配置项的作用：

|        配置项         |                                         说明                                         |
| -------------------- | ------------------------------------------------------------------------------------ |
| debug                | 调试模式开关（true：日志打印和文件存储；false：日志仅文件存储不打印）                       |
| serviceName          | 服务名称                                                                              |
| listenOut             | 是否监听外网访问：1允许，0不允许                                                                          |
| listenPort              | 监听Http访问端口                                                                           |
| callKey       | 服务之间调用的密钥，记住请保持每个服务这个配置相同                                         |
| callRetry       | 服务重试策略，格式`timer1/timer2/timer3/...`（单位毫秒）                                        |
| etcdAddress          | Etcd地址，填写正确的IP加端口，如果是etcd集群的话可以多行填写                       |
| zipkinAddress        | zipkin地址，格式：http://192.168.125.185:9411/api/v2/spans
| amqpAddress        | rabbitmq地址，格式：amqp://guest:guest@192.168.125.185:5672

好了，配置文件创建好了，那么现在再来启动一下程序 `go run main.go` 看看吧！

本以为会开开心心的看到程序启动成功，没想到又报了一个错：

```
PS D:\go-garden\examples\gateway> go run .\main.go
2021/09/18 14:03:41 [Config] Config File "routes" Not Found in "[D:\\go-garden\\examples\\gateway\\configs]"
exit status 1
```

这也是一个配置文件找不到的错误，这次是找不到 `routes` ，这个配置文件是干嘛的呢？

想想看，不管是gateway网关调用下游业务服务还是服务A调用服务B，是不知道下游服务的具体请求地址的，可能只知道他这个接口叫做 `login`，可能完整的地址是 `/api/user/login`
，也可能是 `/api/v1/user/login`，那么现在要调用服务B的 `login`
，就要根据路由配置来获取具体的请求地址。在传统架构里可能直接把地址写在代码里面，万一某一天服务B修改了这个接口的路由，那么得在所有上游服务修改代码更新为正确地址。

好了，言归正传，现在来创建路由配置 `routes.yml` 吧！

```yml
routes:
  user:
    login:
      type: out
      path: /login
      limiter: 5/10000
      fusing: 5/1000
      timeout: 2000
    exists:
      type: in
      path: /exists
      limiter: 5/10000
      fusing: 5/1000
      timeout: 2000
  pay:
    order:
      type: out
      path: /order
      limiter: 5/10000
      fusing: 5/1000
      timeout: 2000
```

第二行 `user` 表示的是user服务，因为go-garden是微服务框架嘛，那么一个项目肯定会拆分到很多的服务，例如 用户中心、支付中心、数据中心等等，这里的 `user` 代表的就是用户中心服务；

`user`下面有两项，分别是 `login` 和 `exists` ，它们表示的是user服务有两个接口，名称分别为 login 和 exists；

每个接口下面包含以下参数：

1、`type`表示接口类型：`out`类型表示面向客户端的接口，只能由`gateway`网关进行调用，其他服务无法调用；`in`表示内部远程调用接口，只能由非`gateway`的其他服务调用；

例如示例中的`user/exists`是提供给`pay`服务进行调用的接口，我们无法在客户端请求`gateway`网关的`api/user/exists`接口。

2、`path`表示请求接口路由：如果服务 `login` 接口路径是 `/api/v1/login` ，那么就要在这里要在这里修改它，`user`
服务监听login接口应该是这样子的： `    r.POST("login", func(c *gin.Context) {...})` ，那么依次类推下面的 `exists` 接口和下面的 `pay` 服务的 `order`
接口都是一样的意思，在这里为了简单就没有写多余的接口路径，大家在实际开发项目中是可以增加 `v1` 这样的前缀的，以免未来更新接口不兼容老接口时可以增加 `v2` 前缀，总之，这个配置文件就是为了实现 `服务->接口名->接口路由`
这个规则。

3、`limiter`表示服务接口限流策略：5/1000表示接口每5秒钟之内最多处理1000个请求，如果超出1000个请求，直接会返回错误响应。

4、`fusing`表示服务熔断策略：`5/100`表示接口每5秒钟之内下游服务器返回了100次错误响应后，直接会对下游服务熔断，在当前5秒内不请求下游服务，直接会返回错误响应。

5、`timeout`表示接口超时控制，单位为毫秒ms，当下游服务接口请求超时将会熔断计数+1且不进行重试。

现在把配置文件创建好后，可以再次启动程序 `go run main.go` ：

```
PS D:\go-garden-demo\gateway> go run .\main.go
2021-09-28T11:55:25.499+0800    INFO    core/gin.go:48  [gateway] Http listen on: 0.0.0.0:8080
2021-09-28T11:55:25.503+0800    INFO    core/amqp.go:80 [amqp] sync consumer is running
```

gateway网关服务启动成功啦！根据打印信息，可以看到服务监听http地址。 现在可以用postman访问网关服务了，地址： `http://127.0.0.1:8080`
，会返回404找不到资源，因为没有路径所以网关找不到请求下游服务的路由配置，带上有路由的路径试试：`http://127.0.0.1:8080/api/user/login`，可以发现网关通过路由配置访问下游服务的格式为 `
/api/服务名称/接口名称`，api前缀是固定的；请求返回404，因为我们并没有启动`user`服务，网关找不到如何请求`user`，所以依然返回404找不到资源：

```json
{
    "msg": "The resource could not be found",
    "status": false
}
```

status是一个bool格式，false说明请求出错了，查看日志信息：

```log
2021-09-28T12:00:07.686+0800    ERROR   core/gateway.go:32      [CallService] service not found
```

看出是调用下游服务 `user`的`login`接口出错， `service not found` 是因为没有启动`user`服务，所以找不到``服务地址，所以根本没法请求到下游的`user`
服务，那么我们下面继续启动`user`服务。

示例地址[examples/gateway](https://github.com/panco95/go-garden/tree/master/examples/gateway)

### 3. User服务

跟gateway服务步骤一样创建好项目`user`和配置文件`config.yml`、`routes.yml`：

创建main.go程序启动入口文件：

```golang
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"sync"
)

var service *core.Garden
var users sync.Map

func main() {
	service = core.New()
	service.Run(route, nil)
}

func route(r *gin.Engine) {
	r.Use(service.CheckCallSafeMiddleware()) // 调用接口安全验证
	r.POST("login", login)
	r.POST("exists", exists)
}

func login(c *gin.Context) {
	var Validate vLogin
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(200, apiResponse(1000, "参数非法", nil))
		return
	}
	username := c.DefaultPostForm("username", "")
	users.Store(username, 1)
	c.JSON(200, apiResponse(0, "登录成功", nil))
}

func exists(c *gin.Context) {
	var Validate vExists
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(200, apiResponse(1000, "参数非法", nil))
		return
	}
	username := c.DefaultPostForm("username", "")
	exists := true
	if _, ok := users.Load(username); !ok {
		exists = false
	}
	c.JSON(200, apiResponse(0, "", core.MapData{
		"exists": exists,
	}))
}

type vLogin struct {
	Username string `form:"username" binding:"required,max=20,min=1"`
}

type vExists struct {
	Username string `form:"username" binding:"required,max=20,min=1"`
}

func apiResponse(code int, msg string, data interface{}) core.MapData {
	return core.MapData{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}
```

示例地址[examples/user](https://github.com/panco95/go-garden/tree/master/examples/user)

观察代码，启动服务的时候有两个参数跟gateway不一样：

1、`service.Run()`第一个参数是路由，因为gateway的路由是在go-garden内部集成的，所以`gateway`服务直接使用了`GatewayRoute`，`user`
服务需要自己实现路由，就是下面的`Route`函数，这是基于`Gin`框架的路由，第一行`r.Use(service.CheckCallSafeMiddleware())`这是校验服务调用安全密钥的中间件，防止非法请求到`user`服务，下面两行`r.Post("login",Login)`和`r.Post("exists",Exists)`就是具体的接口实现，再看看`routes.yml`
可以看出是对应上的，假设这么写路由`r.Post("v1/login",Login)`，那在`routes.yml`应该写成`login: /v1/login`;

2、第二个参数是全局中间件，在`gateway`网关服务中需要实现全局鉴权，所以我们添加了一个`Auth`中间件，我们假设`user`不需要单独的鉴权，所里这里直接写`nil`。

启动`user`服务：`go run main.go`，查看输出：

```
2021-09-28T13:45:50.728+0800    INFO    core/gin.go:48  [user] Http listen on: 192.168.8.98:8081
2021-09-28T13:45:50.732+0800    INFO    core/amqp.go:80 [amqp] sync consumer is running
```

跟`gateway`一样，监听了Http端口；切换到`gateway`服务的窗口，输出了信息：

```
2021-09-28T13:49:58.666+0800    INFO    core/service_manager.go:101     [Service] [user] node [192.168.8.98:8081] join
```

> 表示`gateway`发现了`user`的一个服务节点；

总结一下，这就是go-garden的`服务注册发现`特性，不论你启动多少个服务多少个节点，它们都能互相发现和通信。

现在`user`服务启动成功，现在可以再次使用postman访问`gateway`的`user`
服务路由：`http://127.0.0.1:8080/api/user/login`，增加一个请求参数`username`，发送请求，响应如下：

```json
{
  "code": 0,
  "data": null,
  "msg": "登录成功",
  "status": true
}
```

返回参数`status`为`true`表示请求成功，注意，只有`status`参数是`gateway`返回的，以告诉客户端请求是否成功，其他参数都是`user`返回的，这样使得`gateway`和其他服务无业务耦合度，数据格式可以由服务开发者自行设计。

### 4. 服务集群

go-garden基于`服务自动注册发现`特性，支持大规模的服务集群，例如`user`服务我们可以启动多个示例，现在我们复制一份`user`服务的代码，修改`config.yml`的两个监听端口防止端口冲突，改好端口后启动第二个`user`服务节点：`go run main.go`，查看输出：

```
PS D:\go-garden-demo\user2> go run .\main.go
2021-09-28T13:55:05.284+0800    INFO    core/gin.go:48  [user] Http listen on: 192.168.8.98:8082
2021-09-28T13:55:05.289+0800    INFO    core/amqp.go:80 [amqp] sync consumer is running
```

这是启动的第二个`user`服务节点，`gateway`节点和第一个`user`节点都会`发现`它，切换到`gateway`和`user`节点1窗口，都会输出：

```
2021-09-28T13:55:05.282+0800    INFO    core/service_manager.go:101     [Service] [user] node [192.168.8.98:8082] join
```

现在`user`服务有两个节点，称之为`user`服务集群，那么`gateway`调用`user`服务的时候会是什么一个情况呢？

我们再次使用Postman给`gateway`发送多次请求：`http://127.0.0.1:8080/api/user/login` ，会发现`user`服务节点1和节点2都会打印请求日志，这是`gateway`服务控制的下游服务集群的负载均衡，go-garden默认使用最小连接数算法进行选取下游节点；

在后面的服务之前调用也是会经过负载均衡到下游服务集群。

如果`gateway`也是集群怎么给`gateway`负载均衡呢，建议是增加一层较稳定的`nginx`集群，或者是第三方的负载均衡服务分发到`gateway`节点，更高级的那就是dns的负载均衡，这里不在框架的控制范围，请开发者参考其他文章。

### 5. 服务之间调用

现在解决了`gateay`网关到下游服务的路由分发，那么现在来解决服务到服务之间的调用，`user`服务接口`exists`的`type`为`in`，表示这个接口是对内接口，我们客户端无法通过`gateway`调用它，只能是其他服务调用它，这里我们使用`pay`服务调用；

* 提示：服务之间调用的路由类型为`in`，对外接口为`out`

现在来创建一个`pay`服务，`config.yml`中`ServiceName`改为`pay`，listenPort修改为没有使用过的端口，然后创建main.go程序启动入口文件：

```golang
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"math/rand"
	"time"
)

var service *core.Garden

func main() {
	service = core.New()
	service.Run(route, nil)
}

func route(r *gin.Engine) {
	r.Use(service.CheckCallSafeMiddleware())
	r.POST("order", order)
}

func order(c *gin.Context) {
	span, err := core.GetSpan(c)
	if err != nil {
		c.JSON(500, nil)
		service.Log(core.ErrorLevel, "GetSpan", err)
		return
	}

	var Validate vOrder
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(200, apiResponse(1000, "非法参数", nil))
		return
	}
	username := c.DefaultPostForm("username", "")

	// call [user] service example
	code, result, err := service.CallService(span, "user", "exists", &core.Request{
		Method: "POST",
		Body: core.MapData{
			"username": username,
		},
	})
	if err != nil {
		c.JSON(500, nil)
		service.Log(core.ErrorLevel, "CallService", err)
		span.SetTag("CallService", err)
		return
	}

	var res core.MapData
	err = json.Unmarshal([]byte(result), &res)
	if err != nil {
		c.JSON(500, nil)
		service.Log(core.ErrorLevel, "JsonUnmarshall", err)
		span.SetTag("JsonUnmarshall", err)
	}

	// Parse to get the data returned by the user service, and if the user exists (exists=true), then the order is successful
	data := res["data"].(map[string]interface{})
	exists := data["exists"].(bool)
	if !exists {
		c.JSON(code, apiResponse(1000, "下单失败", nil))
		return
	}
	orderId := fmt.Sprintf("%d%d", time.Now().Unix(), rand.Intn(10000))
	c.JSON(code, apiResponse(0, "下单成功", core.MapData{
		"orderId": orderId,
	}))
}

type vOrder struct {
	Username string `form:"username" binding:"required,max=20,min=1" `
}

// apiResponse format response
func apiResponse(code int, msg string, data interface{}) core.MapData {
	return core.MapData{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}
```

示例代码[examples/pay](https://github.com/panco95/go-garden/tree/master/examples/pay)

观察代码，`pay`服务有一个`order`接口，这是下单接口，请求这个接口需要传参数`username`，用Postman请求一下试试吧！增加`username`请求参数，请求`gateway`网关：`http://127.0.0.1:8080/api/pay/order` ，随便填看看返回什么：

```json
{
  "code": 1000,
  "data": null,
  "msg": "下单失败",
  "status": true
}
```

下单失败了，原因是`order`接口中间会调用`user`服务`exists`接口查询你传入的`username`是否在系统中存在，如果存在才会下单成功，我们观察调用`user`服务核心代码：

* 首先函数开头从请求上下文获取`span`，这是链路追踪相关变量，暂时不管
* 接着获取请求参数`username`
* 下面是调用服务核心函数`service.CallService`，第二个参数表示服务名称，第三个参数表示接口名称，这里我们调用的是`user`服务的`exists`接口，第四个参数表示请求报文，我们定义了参数`username`
* 请求成功后，我们接收到`user`返回的数据，如果返回参数`exists`为true就生成订单号返回成功数据

现在重新请求`http://127.0.0.1:8080/api/user/login`和`http://127.0.0.1/api/pay/order` ，记得两次请求参数`username`
要保持一致，这样业务逻辑才会顺畅，下单接口正确返回数据如下：

```json
{
    "code": 0,
    "data": {
        "orderId": "16328091331318"
    },
    "msg": "下单成功",
    "status": true
}
```

大功告成，服务之间的相互调用就是这么简单！


### 6. 分布式链路追踪

刚刚执行的若干请求，实际上不仅在服务目录生成了日志文件，还在链路追踪服务里记录了一个请求的完整trace日志，也就是客户端请求由`gateway->pay->user`的完整服务执行路径；

准备工作中我们用Docker启动了`Zipkin`，现在要去后台查看数据了，使用浏览器打开`Zipkin`启动服务器IP的9411端口，例如我的虚拟机地址为`http://192.168.125.184:9411`;

打开后就是`Zipkin`的后台界面，我们点击`Run Query`按钮查询，查询结果就出来了，一个`order`接口请求经过三个服务，每个服务的执行时间和相关调试数据都可以查询到，截图：

![pic-1](opentrace-1.png "1")
![pic02](opentrace-2.png "1")

我们在业务代码里也可以非常简单的存储调试数据，go-garden内部已经实现了服务之间的链路关联，代码示例：

```golang
func Test(c *gin.Context) {
span, _ := service.GetSpan(c) //获取span
span.SetTag("key", "value")   //存储数据
}
```


### 7. 动态路由、自动同步、服务配置

1、不管是`gateway`路由分发还是其他服务之间调用，go-garden都是读取`routes.yml`路由配置文件来做相关操作的，假设需要修改/增加/删除一个路由配置，go-garden会监听到`routes.yml`路由配置文件的变化从而更新路由，这是单机服务动态配置；

2、go-garden是分布式的服务框架，并不是一个单机的服务，有各种服务集群在线上运行。那么问题来了，如何保证所有服务集群配置统一呢？ go-garden实现了所有服务之间的`routes.yml`配置文件实时同步，并不需要开发者关心同步逻辑；开发者只需要更新整个架构中任意一个服务的配置文件，就会自动同步到其他服务；

试着修改`gateway`服务的`routes.yml`后保存，然后打开其他服务的配置文件看看，会发现已经同步好了。

3、业务中可使用下面方法获取服务自定义配置项：
* service.GetConfigValue()
* service.GetConfigValueString()
* service.GetConfigValueStringSlice()
* service.GetConfigValueInt()
* service.GetConfigValueIntString()
* service.GetConfigValueMap()

4、自定义配置示例：

```yml
config:
  map:
    a: 1
    b: 2
  int: 1
  intSlice:
    - 1
    - 2
    - 3
  string: hello
  stringSlice:
    - a
    - b
```

分别使用对应的类型获取配置数据。

### 8. 服务限流

在`config.yml`中我们可以给每个服务的每个接口配置单独的限流规则`limiter`参数，`5/1000`表示每5秒钟之内最多处理1000个请求，超出数量不会请求下游服务。

### 9. 服务熔断

在`config.yml`中我们可以给每个服务的每个接口配置单独的熔断规则`fusing`参数，`5/100`表示接口每5秒钟之内下游服务器返回了100次错误响应后，直接会对下游服务熔断，在当前5秒内不请求下游服务，直接会返回错误响应。

### 10. 服务重试

在调用下游服务时，下游服务可能会返回错误，go-garden支持重试机制，在config.yml中配置`callRetry`参数，格式 `timer1/timer2/timer3/...`，可不限制调整，重试次数使用`/`分隔，例如`100/200/200/200/500`表示重试5次，第一次100毫秒，第二次200毫秒，第三次200毫秒，第四次200毫秒，第五次500毫秒，如果重试第五次依然失败，会放弃重试返回错误。大家可根据项目自行调整重试策略配置。

### 11. 超时控制

在调用下游服务时，下游服务可能会超时，go-garden支持超时控制防止超时问题加重导致服务雪崩，在routes.yml中给每个路由配置`timeout`参数，单位为毫秒ms，当下游服务接口请求超时将会熔断计数+1且不进行服务重试。

### 12. 日志

提示：配置文件的`Debug`参数为`true`时，代表调试模式开启，任何日志输出都会同时打印在屏幕上和日志文件中，如果改为`false`，不会在屏幕打印，只会存储在日志文件中

go-garden封装了规范的日志函数，用如下代码进行调用：

```golang
    service.Log(core.ErrorLevel, "JsonUnmarshall", err)
```

第一个参数为日志级别，在源码`core/standard.go`文件中有定义，第二个参部为日志标识，第三个参数为日志内容，建议传入`error`或`string`变量。

### 13、快速构建项目脚手架

访问 [gctl工具](../tools/gctl) 查看使用说明
