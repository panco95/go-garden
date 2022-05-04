# 基于go-garden快速构建微服务

> 提示：go-garden的http服务基于Gin开发，在教程中会涉及到Gin框架的一些内容，例如请求上下文、中间件等，如开发者不了解Gin，请先阅读Gin相关文档！

我们在本教程中会创建一个微服务，包括如下服务：

1、gateway服务，俗称api网关，接收所有接口的客户端请求，然后转发给其他业务服务；

2、user服务，提供login接口保存username，提供exists rpc方法供其他服务查询username用户是否存在；

3、pay服务，提供order接口下单，参数为用户名username，在接口中会rpc调用user的exists方法查询username是否存在，存在下单成功，不存在下单失败。

### 一. 环境准备

go-garden基于Etcd实现服务注册发现，基于Zipkin或Jaeger实现链路追踪，启动必须安装好Etcd、Zipkin或Jaeger

* 在这里给不熟悉的同学介绍Docker快速安装
* 示例环境仅作为测试使用，不可用于生产环境
* zipkin和jaeger都是链路追踪系统，选择一个即可，推荐jaeger，如果不想接入可以删掉相关配置

```sh
docker run -it -d --name etcd -p 2379:2379 -e "ALLOW_NONE_AUTHENTICATION=yes" -e "ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379" bitnami/etcd

docker run -it -d --name zipkin -p 9411:9411 openzipkin/zipkin

docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14250:14250 \
  -p 14268:14268 \
  -p 14269:14269 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.33
```

### 二. 启动Gateway（统一api网关）

安装 [脚手架工具](../tools/garden) 执行命令创建网关服务，服务名称为`my-gateway`：

```sh
garden new my-gateway gateway
```

目录结构请参考文档：[目录结构](../tools/garden#目录结构)

项目创建好后我们需要修改配置文件才能成功启动，修改`configs/config.yml`服务配置文件：

|          字段           |                              说明                               |
| ---------------------- | --------------------------------------------------------------- |
| service->debug         | 调试模式（true：打印、写入文件；false：仅写入文件） |
| service->serviceName   | 服务名称                                                         |
| service->serviceIp     | 服务器内网IP（如服务调用不正常需配置此项)                                                       |
| service->httpOut       | http端口是否允许外网访问：true允许，false不允许                     |
| service->httpPort      | http监听端口                                                     |
| service->allowCors     | http是否允许跨域                                                    |
| service->rpcOut        | rpc端口是否允许外网访问：true允许，false不允许                     |
| service->rpcPort       | rpc监听端口                                                     |
| service->callKey       | 服务之间调用的密钥，请保持每个服务一致                              |
| service->callRetry     | 服务重试策略，格式`timer1/timer2/timer3/...`（单位毫秒）           |
| service->etcdKey       | Etcd关联密钥，一套服务使用同一个key才能实现服务注册发现              |
| service->etcdAddress   | Etcd地址，填写正确的IP加端口，如果是etcd集群的话可以多行填写         |
| service->tracerDrive   | 分布式链路追踪引擎，可选zipkin、jaeger，如果不需要，删掉此项配置      |
| service->zipkinAddress | zipkin上报地址，格式：http://127.0.0.1:9411/api/v2/spans       |
| service->jaegerAddress | jaeger上报地址，格式：127.0.0.1:6831       |
| service->pushGatewayAddress | 服务监控Prometheus->pushGateway上报地址，格式：127.0.0.1:9091       |
| config->*              | 自定义配置项                                           |

修改好对应的配置后，启动服务：

* 启动服务有两个参数可选，configs为指定配置文件目录，runtime为指定日志输出目录，默认为当前路径configs目录和runtime目录

```sh
go run main.go -configs=configs -runtime=runtime
```

成功输出：

```sh
{"level":"info","time":"2022-05-03 18:17:10","caller":"core/bootstrap.go:15","msg":"[bootstrap] my-gateway running"}
{"level":"info","time":"2022-05-03 18:17:10","caller":"core/opentracing.go:37","msg":"loggingTracer created","sampler":"ConstSampler(decision=true)","tags":[{"Key":"jaeger.version","Value":"Go-2.30.0"},{"Key":"hostname","Value":"localhost.localdomain"},{"Key":"ip","Value":"192.168.129.151"}]}
{"level":"info","time":"2022-05-03 18:17:10","caller":"core/rpc.go:26","msg":"[rpc] listen on: 192.168.129.151:9000"}
{"level":"info","time":"2022-05-03 18:17:10","caller":"core/gin.go:61","msg":"[http] listen on: 0.0.0.0:8080"}
{"level":"info","time":"2022-05-03 18:17:10","caller":"server/server.go:198","msg":"server pid:44373"}
```


注意：如果出现错误请查看go.mod中go-garden包版本号是否为最新版！

### 三. 启动User服务

执行命令创建user服务，服务名称为`my-user`：

```sh
garden new my-user service
```

同样修改`configs/config.yml`配置文件，如果跟gateway在同一台主机，需要修改httpPort和rpcPort防止端口冲突；启动服务：

```sh
go run main.go -configs=configs -runtime=runtime
```

这时gateway会发现user服务节点加入且输出信息节点信息。

### 四. 定义user路由

路由文件路径为`configs/routes.yml`，我们需要正确修改路由文件方才能让框架内部正常执行请求链路；修改`routes.yml`：

```yml
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

| 字段 | 说明 |
| ---------------------- | --------------------------------------------------------------- | 
| my-user | 服务名称 | | my-user->login | my-user服务的login路由配置 |
| my-user->login->type | 路由类型：http，表示此接口是api接口，由gateway调用转发 |
| my-user->login->path | http路由类型时需要此配置，表示login接口完整路由 |
| my-user->login->limiter | 服务限流器，5/100表示login接口5秒内最多接受100个请求，超出后限流 |
| my-user->login->fusing | 服务熔断器，5/100表示login接口5秒内最多允许100次错误，超出后熔断 |
| my-user->login->timeout | 服务超时控制，单位ms，2000表示请求login接口超出2秒后不等待结果 |
| my-user->exists | my-user服务的exists路由配置 |
| my-user->exists->type | 路由类型：rpc，表示此接口是rpc方法，由业务服务之间调用 |
| my-user->exists->limiter | 服务限流器，5/100表示exists方法5秒内最多接受100个请求，超出后限流 |
| my-user->exists->fusing | 服务熔断器，5/100表示exists方法5秒内最多允许100次错误，超出后熔断 |
| my-user->exists->timeout | 服务超时控制，单位ms，2000表示请求exists方法超出2秒后不等待结果 |

修改好路由配置后保存，框架会热更新路由配置且同步到其他的服务，无需重启服务；可以观察`my-gateway`的路由配置文件已经同步为`my-user`的路由配置文件了。

### 五. 编写user服务api接口

上面我们定义了user服务的login接口，现在我们来实现它；

创建全局变量Users（简单代替mysql数据库存储），用于保存用户信息，`global/global.go`：

```go
package global

import (
	"github.com/panco95/go-garden/core"
	"sync"
)

var (
	Garden *core.Garden
	Users  sync.Map
)
```

创建`api/login.go`，编写login接口代码：

```go
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
		Fail(c, MsgInvalidParams)
		return
	}
	username := c.PostForm("username")
	global.Users.Store(username, 1)
	Success(c, MsgOk, nil)
}
```

添加login接口路由path定义，`api/base.go`：

```go
package api

import (
	"github.com/gin-gonic/gin"
	"my-user/global"
)

func Routes(r *gin.Engine) {
	r.POST("login", Login)
}
```

### 六：访问api接口

实现了user服务的login接口后，现在通过客户端来请求它；

重启user服务，打开`postman`或其他接口测试工具；

请求地址格式：http://[gateway地址]:[gateway http端口]/api/[服务名称]/[服务接口]/[接口path] ；

所以login接口完整的请求地址为：`http://127.0.0.1:8080/api/my-user/login`；

修改请求类型为post，增加请求参数username，发出请求：

```json
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

现在我们要给user服务增加一个exists方法给其他服务调用，首先在`rpc/define`定义exists方法的调用参数和返回参数，`rpc/define/exists.go`：

```go
package define

type ExistsArgs struct {
	Username string
}

type ExistsReply struct {
	Exists bool
}
```

ExistsArgs是调用的参数结构体，ExistsReply是方法返回的结构体；

接着在增加方法具体逻辑，`rpc/exists.go`：

```go
package rpc

import (
	"context"
	"my-user/global"
	"my-user/rpc/define"
)

func (r *Rpc) Exists(ctx context.Context, args *define.ExistsArgs, reply *define.ExistsReply) error {
	span := global.Garden.StartRpcTrace(ctx, args, "Exists")

	reply.Exists = false
	if _, ok := global.Users.Load(args.Username); ok {
		reply.Exists = true
	}

	global.Garden.FinishRpcTrace(span)
	return nil
}
```

重启user服务，rpc方法`exists`就写好了。

### 八：调用user服务rpc方法

增加一个pay服务，在其api里来调用user的rpc方法，创建pay服务，服务名称为`my-pay`：

```sh
garden new my-pay service
```

修改配置文件`configs/config.yml`，然后启动服务：

```sh
go run main.go -configs=configs -runtime=runtime
```

启动成功后我们把定义一下路由文件，然后在服务开启状态会自动同步给其他服务，`configs/routes.yml`：

```yml
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
  my-pay:
    order:
      type: http
      path: /order
      limiter: 5/100
      fusing: 5/100
      timeout: 2000
```

我们给pay服务增加了一个order接口，我们在order接口实现里调用user服务的exists rpc方法；

首先我们把exists方法的rpc参数定义，就类似grpc的protobuf，把user服务那里定义的赋值进来就好，`rpc/user/exists.go`：

```go
package user

type ExistsArgs struct {
	Username string
}

type ExistsReply struct {
	Exists bool
}
```

编写接口order接口业务逻辑，`rpc/order.go`：

```go
package rpc

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"math/rand"
	"my-pay/global"
	"my-pay/rpc/user"
	"time"
)

func Order(c *gin.Context) {
	var validate struct {
		Username string `form:"username" binding:"required,max=20,min=1" `
	}
	if err := c.ShouldBind(&validate); err != nil {
		Fail(c, MsgInvalidParams)
		return
	}
	username := c.DefaultPostForm("username", "")

	span := core.GetSpan(c)
	args := user.ExistsArgs{
		Username: username,
	}
	reply := user.ExistsReply{}
	err = global.Garden.CallRpc(span, "user", "exists", &args, &reply)
	if err != nil {
		Fail(c, MsgFail)
		global.Garden.Log(core.ErrorLevel, "rpcCall", err)
		span.SetTag("callRpc", err)
		return
	}
	if !reply.Exists {
		Fail(c, MsgOk)
		return
	}

	orderId := fmt.Sprintf("%d%d", time.Now().Unix(), rand.Intn(10000))
	Success(c, MsgOk, core.MapData{
		"orderId": orderId,
	})
}
```

接着修改api请求路由，`api/base.go`：

```go
package api

import (
	"github.com/gin-gonic/gin"
	"my-pay/global"
)

func Routes(r *gin.Engine) {
	r.POST("order", Order)
}
```

上面就是rpc方法调用的全部步骤，跟grpc很像，只不过go-garden没有使用protobuf编码协议，不需要定义protobuf文件然后用工具执行编码、解码，只需要定义结构体即可，非常方便。

重启pay服务，现在来测试一下rpc调用是否正常，通过gateway访问pay服务的order方法完整地址为： `http://127.0.0.1:8080/api/my-pay/order` ，带上跟user/login一样的参数请求，返回响应：

```json
{
  "code": 0,
  "data": {
    "orderId": "16353190617887"
  },
  "msg": "下单成功",
  "status": true
}
```

如果返回的是正确响应，说明rpc调用成功了，如果返回的是下单失败，是因为你调用接口传的username参数跟用户服务login接口传的不一致导致的。

### 九. 分布式链路追踪

我们刚刚请求pay服务的order接口，实际上这个请求经过了三个服务，流程为：

`client->gateway->pay->user`

如果接口突然响应异常，我们如何定位报错位置呢？

第一个方法可以日志排查，在所有调用链的服务的runtime日志找出错误，逐一排查，这种方法在只有2-3个服务的时候勉强行得通，但是也非常的低效；

go-garden内部集成了分布式链路追踪系统，支持zipkin和jaeger，需要在配置文件中配置；

调用链每一层我们都可以记录信息，然后在非常清晰的ui界面上查看，zipkin地址： http://127.0.0.1:9411/zipkin    jaeger地址： http://127.0.0.1:16686

记录链路日志，日志数据支持string和err类型：

```golang
span := core.GetSpan(c)
span.SetTag("key", "val")
span.SetTag("key", err)
```

### 十. 自定义配置

我们在业务中会自定义一些配置，例如您需要在业务中连接mysql、redis等，可在此处自行添加配置项然后通过viper的函数获取配置值，`configs/config.yml`：

```yml
service:
  #service是框架定义的配置项，请不要再后面覆盖service配置项
  #...

mysql:
  enable: true
  addr: 127.0.0.1:3306
  user: root
  pass: abcdefg
  pool: 10
  advanced:
    timeout: 10s
number: 1.111
```

框架已经把配置文件注入到viper，业务中使用viper提供的方法即可快捷获取配置，更多方法请参考viper文档或源码：

* viper.GetString("mysql.addr")
* viper.GetInt("mysql.pool")
* viper.GetBool("mysql.enable")
* viper.GetDuration("mysql.advanced.timeout")
* viper.GetFloat64("number)

### 小插曲：容器全局变量
框架提供了一个全局容器提供给大家使用，如果你觉得全局变量的方式不够优雅，可以用框架提供的容器存储依赖，使用Get()和Set()方法存储/取出依赖：
```golang
err := global.Garden.Set("key", interface{})
if err != nil {
	
}
res, err := global.Garden.Get("key")
if err != nil {
	
}
global.Garden.Log(core.DebugLevel,"container test", res)
```
* value可以存储任意类型，get获取到interface{}类型，需要自行断言类型；
* Get和Set都是并发安全的；

### 十一、负载均衡

上面的每一个服务都只启动了一个节点，同一份代码我们可以在多台服务器上启动，serviceName就是每个服务的标识，同名服务我们就称为服务集群； 复制一份user服务代码修改监听端口，启动；

现在user服务就是两个节点在运行，这时候我们调用user服务接口或者rpc方法的时候，go-garden内部会通过最小连接数以及轮询策略来选择服务器节点进行请求，开发者无需关心内部逻辑。

### 十二. 服务限流

在`config.yml`中我们可以给每个服务的每个接口配置单独的限流规则`limiter`参数，`5/1000`表示每5秒钟之内最多处理1000个请求，超出数量不会请求下游服务。

### 十三. 服务熔断

在`config.yml`中我们可以给每个服务的每个接口配置单独的熔断规则`fusing`参数，`5/100`表示接口每5秒钟之内下游服务器返回了100次错误响应后，直接会对下游服务熔断，在当前5秒内不请求下游服务，直接会返回错误响应。

### 十四. 服务重试

在调用下游服务时，下游服务可能会返回错误，go-garden支持重试机制，在config.yml中配置`callRetry`参数，格式 `timer1/timer2/timer3/...`，可不限制调整，重试次数使用`/`
分隔，例如`100/200/200/200/500`表示重试5次，第一次100毫秒，第二次200毫秒，第三次200毫秒，第四次200毫秒，第五次500毫秒，如果重试第五次依然失败，会放弃重试返回错误。大家可根据项目自行调整重试策略配置。

### 十五. 超时控制

在调用下游服务时，下游服务可能会超时，go-garden支持超时控制防止超时问题加重导致服务雪崩，在routes.yml中给每个路由配置`timeout`参数，单位为毫秒ms，当下游服务接口请求超时将会熔断计数+1且不进行服务重试。

### 十六. 日志

提示：配置文件的`Debug`参数为`true`时，代表调试模式开启，任何日志输出都会同时打印在屏幕上和日志文件中，如果改为`false`，不会在屏幕打印，只会存储在日志文件中

指定日志存储路径：启动时添加参数，例如 `runtime=test/logs`

* garden.log为日志存储文件，自动分割，最大2m
* gin.log是gin日志，开启调试模式才会存储

框架的初始化了log包，请引入 `github.com/panco95/go-garden/core/log` 包进行日志输出：

```go
import "github.com/panco95/go-garden/core/log"

log.Info("label", "log")
log.Infof("label", "log %s", "test")
log.Error(...)
log.Errorf(...)
log.Panic(...)
log.Panicf(...)
log.Error(...)
log.Errorf(...)
log.Warn(...)
log.Warnf(...)
log.Fatal(...)
log.Fatal(...)
```

第一个参数为日志标识，第二个参数为日志内容，支持`error`或`string`。

deply目录下有filebeat同步日志到elasticsearch配置，可根据需要自行使用集成到elasticsearch甚至是grafana：
```shell
filebeat -e -c ./deply/filebeat.yml
```
### 十七. 服务监控与警报

1、集成pprof性能监控，路由：/debug/pprof（需要开启debug调试模式，因为安全问题不建议生产环境开启）

2、支持[Prometheus](https://prometheus.io)：

/metrics接口提供采集指标，在promtheus配置文件中增加接口地址：

prometheus.yml：
```yml
scrape_configs:
  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.
  - job_name: "gateway"
    scrape_interval: 5s
    static_configs:
      - targets: ["192.168.125.193:8080"]

  - job_name: "pay"
    scrape_interval: 5s
    static_configs:
      - targets: ["192.168.125.193:8082"]

  - job_name: "user-1"
    scrape_interval: 5s
    static_configs:
      - targets: ["192.168.125.193:8081"]

  - job_name: "user-2"
    scrape_interval: 5s
    static_configs:
      - targets: ["192.168.125.193:8083"]
```

指标默认为golang_client组件提供，可前往grafana官网搜索metrics模板；

同时支持指标主动上报[PushGateway](https://github.com/prometheus/pushgateway) ，在configs.yml配置好pushGateway地址后可在代码中调用上报：
```golang
data := core.MapData{
	"metric-1": 100,
	"metric-2": 200,
}
global.Garden.PushGateway("jobname", data)
```

按照服务目标和指标名称通过Prometheus查询表达式进行查询：
```shell
RequestProcess{instance="192.168.125.193:8080",job="gateway"}
RequestFinish{instance="192.168.125.193:8080",job="gateway"}
RequestFinish{job="gateway"}
RequestFinish{job="user-1"}
RequestFinish
```

* 推荐搭配[grafana]一起使用(https://grafana.com/)


### 十八. Docker部署

deply目录里面有最佳实践Dockerfile和docker-compose.yml文件，仅供参考！

```dockerfile
FROM golang:1.17 as mod
LABEL stage=mod
ARG GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,https://goproxy.io,direct
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

FROM mod as builder
LABEL stage=intermediate0
ARG LDFLAGS
ARG GOARCH=amd64
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} \
   go build -o main \
   -gcflags="all=-trimpath=`pwd` -N -l" \
   -asmflags "all=-trimpath=`pwd`" \
   -ldflags "${LDFLAGS}" main.go


FROM alpine:3.13.5

LABEL MAINTAINER="panco 1129443982@qq.com" \
    URL="https://github.com/panco95"

COPY --from=builder /app/main /main

ENV TZ Asia/Shanghai

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache \
      curl \
      ca-certificates \
      bash \
      iproute2 \
      tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo Asia/Shanghai > /etc/timezone && \
    if [ ! -e /etc/nsswitch.conf ];then echo 'hosts: files dns myhostname' > /etc/nsswitch.conf; fi && \
   rm -rf /var/cache/apk/* /tmp/*

ENTRYPOINT ["/main"]
```

```
version: '3'

services:

  gateway:
    build: .
    container_name: gateway-1
    ports:
      - "8080:8080"
      - "9000:9000"
    volumes:
      - /etc/gateway:/configs
      - /var/log/gateway:/runtime
```