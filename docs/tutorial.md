# 基于go-garden快速构建微服务

> 提示：go-garden的http服务基于Gin开发，在教程中会涉及到Gin框架的一些内容，例如请求上下文、中间件等，如开发者不了解Gin，请先阅读Gin相关文档！

### 一. 环境准备
go-garden基于Etcd实现服务注册发现，基于Zipkin实现链路追踪，启动必须安装好Etcd、Zipkin

* 在这里给不熟悉的同学介绍Docker快速安装
* 示例环境仅作为测试使用，不可用于生产环境
 
```
docker run -it -d --name etcd -p 2379:2379 -e "ALLOW_NONE_AUTHENTICATION=yes" -e "ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379" bitnami/etcd
docker run -it -d --name zipkin -p 9411:9411 openzipkin/zipkin
```

### 二. 启动Gateway（统一api网关）

安装 [脚手架工具](../tools/garden) 执行命令创建网关服务，服务名称为my-gateway：
```
garden new my-gateway gateway
```
项目创建好后我们需要修改配置文件才能成功启动，修改`configs/config.yml`服务配置文件：

|          字段           |                              说明                               |
| ---------------------- | --------------------------------------------------------------- |
| service->debug         | 调试模式开关（true：日志打印和文件存储；false：日志仅文件存储不打印） |
| service->serviceName   | 服务名称                                                         |
| service->httpOut       | http端口是否允许外网访问：true允许，false不允许                     |
| service->httpPort      | http监听端口                                                     |
| service->rpcOut        | rpc端口是否允许外网访问：true允许，false不允许                     |
| service->rpcPort       | rpc监听端口                                                     |
| service->callKey       | 服务之间调用的密钥，请保持每个服务一致                              |
| service->callRetry     | 服务重试策略，格式`timer1/timer2/timer3/...`（单位毫秒）           |
| service->etcdKey       | Etcd关联密钥，一套服务使用同一个key才能实现服务注册发现              |
| service->etcdAddress   | Etcd地址，填写正确的IP加端口，如果是etcd集群的话可以多行填写         |
| service->zipkinAddress | zipkin地址，格式：http://192.168.125.185:9411/api/v2/spans       |
| config->*              | 自定义配置项，后面有说明                                          |

修改好对应的配置后，启动服务：

```
go run main.go
```

启动成功输出：
```
2021-10-27 09:49:18     info    core/bootstrap.go:9     [bootstrap] my-gateway service starting now...
2021-10-27 09:49:18     info    core/rpc.go:16  [rpc] listen on: 192.168.8.98:9000
2021-10-27 09:49:18     info    core/gin.go:49  [http] listen on: 0.0.0.0:8080
2021/10/27 09:49:18 server.go:198: INFO : server pid:16224
```

### 三. 启动User服务
执行命令创建user服务，服务名称为my-user：
```
garden new my-user service
```
同样修改`configs/config.yml`配置文件，如果跟gateway在同一台主机，需要修改httpPort和rpcPort防止端口冲突；启动服务：
```
go run main.go
```
启动成功输出：
```
2021-10-27 09:50:08     info    core/bootstrap.go:9     [bootstrap] my-user service starting now...
2021-10-27 09:50:08     info    core/rpc.go:16  [rpc] listen on: 192.168.8.98:9001
2021-10-27 09:50:08     info    core/gin.go:49  [http] listen on: 0.0.0.0:8081
2021/10/27 09:50:08 server.go:198: INFO : server pid:23740
```

这时gateway发现user服务节点加入且输出信息：
```
2021-09-28T13:49:58.666+0800    INFO    core/service_manager.go:101     [Service] [user] node [192.168.8.98:8081] join
```

### 四. 定义user路由
路由文件路径为`configs/routes.yml`，我们需要正确修改路由文件方才能让框架内部正常执行请求链路；修改`routes.yml`：
```
routes:
  my-user:
    login:
      type: http
      path: /login
      limiter: 5/100
      fusing: 5/100
      timeout: 2000
    exists:
      type: rpc
      limiter: 5/100
      fusing: 5/100
      timeout: 2000
```
路由说明：
|         字段          |                              说明                               |
| ---------------------- | --------------------------------------------------------------- |
| my-user                           | 服务名称 |
| my-user->login                | my-user服务的login路由配置                                                       |
| my-user->login->type      | 路由类型：http，表示此接口是api接口，由gateway调用转发                    |
| my-user->login->path      | http路由类型时需要此配置，表示login接口完整路由                                                     |
| my-user->login->limiter    |  服务限流器，5/100表示login接口5秒内最多接受100个请求，超出后限流                    |
| my-user->login->fusing    | 服务熔断器，5/100表示login接口5秒内最多允许100次错误，超出后熔断                                                     |
| my-user->login->timeout  | 服务超时控制，单位ms，2000表示请求login接口超出2秒后不等待结果                             |
| my-user->exists                  | my-user服务的exists路由配置                            |
| my-user->exists->type  |     路由类型：http，表示此接口是rpc方法，由业务服务之间调用                      |
| my-user->exists->limiter    |  服务限流器，5/100表示exists方法5秒内最多接受100个请求，超出后限流                    |
| my-user->exists->fusing    | 服务熔断器，5/100表示exists方法5秒内最多允许100次错误，超出后熔断                                                     |
| my-user->exists->timeout  | 服务超时控制，单位ms，2000表示请求exists方法超出2秒后不等待结果                             |

修改好路由配置后保存，框架会热更新路由配置且同步到其他的服务，无需重启服务；可以观察my-gateway的路由配置文件已经同步为my-user的路由配置文件了。

### 五. 编写user服务api接口
上面我们定义了user服务的login接口，现在我们来实现它；
创建全局变量Users（简单代替mysql数据库存储），用于保存用户信息，`global/global.go`：
```
package global

import (
	"github.com/panco95/go-garden/core"
	"sync"
)

var (
	Service *core.Garden
	Users   sync.Map
)
```
创建`api/login.go`，编写login接口代码：
```
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"my-user/global"
)

func Login(c *gin.Context) {
	var validate struct {
		Username string `form:"username" binding:"required,max=20,min=1"`
	}
	if err := c.ShouldBind(&validate); err != nil {
		core.Resp(c, core.HttpOk, -1, core.InfoInvalidParam, nil)
		return
	}
	username := c.DefaultPostForm("username", "")
	global.Users.Store(username, 1)
	core.Resp(c, core.HttpOk, 0, "登陆成功", nil)
}
```
添加login接口路由path定义，`api/base.go`：
```
package api

import (
	"github.com/gin-gonic/gin"
	"my-user/global"
)

func Routes(r *gin.Engine) {
	r.Use(global.Service.CheckCallSafeMiddleware())
	r.POST("login", Login)
}
```
#### 六：访问api接口
实现了user服务的login接口后，现在通过客户端来请求它；
重启user服务，打开`postman`或其他接口测试工具；
请求地址格式：http://[gateway地址]:[gateway http端口]/api/[服务名称]/[服务接口]/[接口path] ；
所以login接口完整的请求地址为：`http://127.0.0.1:8080/api/my-user/login`；
修改请求类型为post，增加请求参数username，发出请求：
```
{
    "code": 0,
    "data": null,
    "msg": "登陆成功",
    "status": true
}
```
gateway服务会通过请求路径，把对应的请求转发给my-user服务，然后my-user返回响应给gateway，gateway接收到my-user的响应内容，返回给客户端；
gateway会把收到的请求结果增加一个status字段，如果请求my-user服务失败会返回false，成功既true。

### 七：编写user服务rpc方法

### 八：调用user服务rpc方法

### 九. 分布式链路追踪

### 十. 自定义配置

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

### 十一. 服务限流

在`config.yml`中我们可以给每个服务的每个接口配置单独的限流规则`limiter`参数，`5/1000`表示每5秒钟之内最多处理1000个请求，超出数量不会请求下游服务。

### 十二. 服务熔断

在`config.yml`中我们可以给每个服务的每个接口配置单独的熔断规则`fusing`参数，`5/100`表示接口每5秒钟之内下游服务器返回了100次错误响应后，直接会对下游服务熔断，在当前5秒内不请求下游服务，直接会返回错误响应。

### 十三. 服务重试

在调用下游服务时，下游服务可能会返回错误，go-garden支持重试机制，在config.yml中配置`callRetry`参数，格式 `timer1/timer2/timer3/...`，可不限制调整，重试次数使用`/`分隔，例如`100/200/200/200/500`表示重试5次，第一次100毫秒，第二次200毫秒，第三次200毫秒，第四次200毫秒，第五次500毫秒，如果重试第五次依然失败，会放弃重试返回错误。大家可根据项目自行调整重试策略配置。

### 十四. 超时控制

在调用下游服务时，下游服务可能会超时，go-garden支持超时控制防止超时问题加重导致服务雪崩，在routes.yml中给每个路由配置`timeout`参数，单位为毫秒ms，当下游服务接口请求超时将会熔断计数+1且不进行服务重试。

### 十五. 日志

提示：配置文件的`Debug`参数为`true`时，代表调试模式开启，任何日志输出都会同时打印在屏幕上和日志文件中，如果改为`false`，不会在屏幕打印，只会存储在日志文件中

go-garden封装了规范的日志函数，用如下代码进行调用：

```golang
    service.Log(core.ErrorLevel, "JsonUnmarshall", err)
```

第一个参数为日志级别，在源码`core/standard.go`文件中有定义，第二个参部为日志标识，第三个参数为日志内容，建议传入`error`或`string`变量。

### 十六、快速构建项目脚手架

访问 [脚手架工具](../tools/garden) 查看使用说明
