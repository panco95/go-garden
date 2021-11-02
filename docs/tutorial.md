# 基于go-garden快速构建微服务

> 提示：go-garden的http服务基于Gin开发，在教程中会涉及到Gin框架的一些内容，例如请求上下文、中间件等，如开发者不了解Gin，请先阅读Gin相关文档！

我们在本教程中会创建一个微服务，包括如下服务：

1、gateway服务，俗称api网关，接收所有接口的客户端请求，然后转发给其他业务服务；

2、user服务，提供login接口保存username，提供exists rpc方法供其他服务查询username用户是否存在；

3、pay服务，提供order接口下单，参数为用户名username，在接口中会rpc调用user的exists方法查询username是否存在，存在下单成功，不存在下单失败。

访问 [examples](../examples) 查看当前教程完整代码

### 一. 环境准备
go-garden基于Etcd实现服务注册发现，基于Zipkin实现链路追踪，启动必须安装好Etcd、Zipkin

* 在这里给不熟悉的同学介绍Docker快速安装
* 示例环境仅作为测试使用，不可用于生产环境
 
```sh
docker run -it -d --name etcd -p 2379:2379 -e "ALLOW_NONE_AUTHENTICATION=yes" -e "ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379" bitnami/etcd
docker run -it -d --name zipkin -p 9411:9411 openzipkin/zipkin
```

### 二. 启动Gateway（统一api网关）

安装 [脚手架工具](../tools/garden) 执行命令创建网关服务，服务名称为`my-gateway`：
```sh
garden new my-gateway gateway
```
项目创建好后我们需要修改配置文件才能成功启动，修改`configs/config.yml`服务配置文件：

|          字段           |                              说明                               |
| ---------------------- | --------------------------------------------------------------- |
| service->debug         | 调试模式开关（true：日志打印和文件存储；false：日志仅文件存储不打印） |
| service->serviceName   | 服务名称                                                         |
| service->httpOut       | http端口是否允许外网访问：true允许，false不允许                     |
| service->httpPort      | http监听端口                                                     |
| service->allowCors      | http是否允许跨域                                                    |
| service->rpcOut        | rpc端口是否允许外网访问：true允许，false不允许                     |
| service->rpcPort       | rpc监听端口                                                     |
| service->callKey       | 服务之间调用的密钥，请保持每个服务一致                              |
| service->callRetry     | 服务重试策略，格式`timer1/timer2/timer3/...`（单位毫秒）           |
| service->etcdKey       | Etcd关联密钥，一套服务使用同一个key才能实现服务注册发现              |
| service->etcdAddress   | Etcd地址，填写正确的IP加端口，如果是etcd集群的话可以多行填写         |
| service->zipkinAddress | zipkin地址，格式：http://192.168.125.185:9411/api/v2/spans       |
| config->*              | 自定义配置项，框架默认定义好redis和数据库配置                                           |

修改好对应的配置后，启动服务：

```sh
go run main.go
```

启动成功输出：
```sh
2021-10-27 09:49:18     info    core/bootstrap.go:9     [bootstrap] my-gateway service starting now...
2021-10-27 09:49:18     info    core/rpc.go:16  [rpc] listen on: 192.168.8.98:9000
2021-10-27 09:49:18     info    core/gin.go:49  [http] listen on: 0.0.0.0:8080
2021/10/27 09:49:18 server.go:198: INFO : server pid:16224
```

### 三. 启动User服务
执行命令创建user服务，服务名称为`my-user`：
```sh
garden new my-user service
```
同样修改`configs/config.yml`配置文件，如果跟gateway在同一台主机，需要修改httpPort和rpcPort防止端口冲突；启动服务：
```sh
go run main.go
```
启动成功输出：
```sh
2021-10-27 09:50:08     info    core/bootstrap.go:9     [bootstrap] my-user service starting now...
2021-10-27 09:50:08     info    core/rpc.go:16  [rpc] listen on: 192.168.8.98:9001
2021-10-27 09:50:08     info    core/gin.go:49  [http] listen on: 0.0.0.0:8081
2021/10/27 09:50:08 server.go:198: INFO : server pid:23740
```

这时gateway发现user服务节点加入且输出信息：
```sh
2021-10-27 11:35:08     info    core/service_manager.go:106     [Service] [my-user] node [192.168.8.98:8081:9001] join
```

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
| my-user->exists->type  |     路由类型：rpc，表示此接口是rpc方法，由业务服务之间调用                      |
| my-user->exists->limiter    |  服务限流器，5/100表示exists方法5秒内最多接受100个请求，超出后限流                    |
| my-user->exists->fusing    | 服务熔断器，5/100表示exists方法5秒内最多允许100次错误，超出后熔断                                                     |
| my-user->exists->timeout  | 服务超时控制，单位ms，2000表示请求exists方法超出2秒后不等待结果                             |

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
	Service *core.Garden
	Users   sync.Map
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
		core.Resp(c, core.HttpOk, -1, core.InfoInvalidParam, nil)
		return
	}
	username := c.DefaultPostForm("username", "")
	global.Users.Store(username, 1)
	core.Resp(c, core.HttpOk, 0, "登陆成功", nil)
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
	r.Use(global.Service.CheckCallSafeMiddleware())
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

接着在增加方法具体萝莉，`rpc/exists.go`：
```go
package rpc

import (
	"context"
	"my-user/global"
	"my-user/rpc/define"
)

func (r *Rpc) Exists(ctx context.Context, args *define.ExistsArgs, reply *define.ExistsReply) error {
   // rpc方法链路追踪开启
	span := global.Service.StartRpcTrace(ctx, args, "Exists")

   // rpc方法具体业务逻辑
	reply.Exists = false
	if _, ok := global.Users.Load(args.Username); ok {
		reply.Exists = true
	}

   // rpc方法链路追踪结束
	global.Service.FinishRpcTrace(span)
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
go run main.go
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
		core.Resp(c, core.HttpOk, -1, core.InfoInvalidParam, nil)
		return
	}
	username := c.DefaultPostForm("username", "")

	span, err := core.GetSpan(c)
	if err != nil {
		core.Resp(c, core.HttpFail, -1, core.InfoServerError, nil)
		global.Service.Log(core.ErrorLevel, "GetSpan", err)
		return
	}

   // 定义rpc exists方法的调用参数和返回参数结构体
	args := user.ExistsArgs{
		Username: username,
	}
	reply := user.ExistsReply{}
	// 调用rpc服务
	err = global.Service.CallRpc(span, "my-user", "exists", &args, &reply)
	if err != nil {
		core.Resp(c, core.HttpFail, -1, core.InfoServerError, nil)
		global.Service.Log(core.ErrorLevel, "rpcCall", err)
		span.SetTag("callRpc", err)
		return
	}
	// 收到exists方法返回数据，判断数据(业务逻辑)
	if !reply.Exists {
		core.Resp(c, core.HttpOk, -1, "下单失败", nil)
		return
	}

	orderId := fmt.Sprintf("%d%d", time.Now().Unix(), rand.Intn(10000))
	core.Resp(c, core.HttpOk, 0, "下单成功", core.MapData{
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
	r.Use(global.Service.CheckCallSafeMiddleware())
	r.POST("order", Order)
}
```

上面就是rpc方法调用的全部步骤，跟grpc很像，只不过go-garden没有使用protobuf编码协议，不需要定义protobuf文件然后用工具执行编码、解码，只需要定义结构体即可，非常方便。

重启pay服务，现在来测试一下rpc调用是否正常，通过gateway访问pay服务的order方法完整地址为：`http://127.0.0.1:8080/api/my-pay/order`，带上跟user/login一样的参数请求，返回响应：

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

go-garden内部集成了分布式链路追踪系统，调用链每一层我们都可以记录信息，然后在非常清晰的ui界面上查看，我们访问zipkin所在服务器网址：`http://127.0.0.1:9411/zipkin/`，可以查询到刚刚的请求链路追踪记录，如下图所示：
![链路追踪zipkin](zipkin.png)


### 十. 自定义配置

我们在业务中会自定义一些配置，如框架集成的数据库、redis，开发者可在此处自行添加配置项，`configs/config.yml`：
```yml
service:

config:
  mysql_open: false
  mysql_addr: "127.0.0.1:3306"
  mysql_user: "root"
  mysql_pass: ""
  mysql_dbname: "test"
  mysql_charset: "utf8mb4"
  mysql_parseTime: true
  mysql_connPool: 10

  redis_open: false
  redis_addr: "127.0.0.1:6379"
  redis_pass: ""
  redis_db: 0

  a: 1.13
  b: 1
  c: "hello"
  d: true
  e:
    e1: "aa"
    e2: 1
```

业务中使用下面方法获取服务自定义配置项：

* 获取map类型配置：GetConfigValueMap("e")
* 获取int类型配置：GetConfigValueInt("b")
* 获取float32类型配置：GetConfigValueFloat32("a")
* 获取float64类型配置：GetConfigValueFloat64("a")
* 获取string类型配置：GetConfigValueString("c")
* 获取bool类型配置：GetConfigValueString("d")
* 获取interface类型配置：GetConfigValueInterface("a").(float64)

### 十一、数据库

框架继承了数据库组件gorm，如需使用请在configs.yml增加如下配置：
```yml
service:
  ---

config:
  mysql_open: true              #是否使用mysql
  mysql_addr: "127.0.0.1:3306"  #mysql连接地址
  mysql_user: "root"            #数据库用户名
  mysql_pass: ""                #数据库密码
  mysql_dbname: "test"          #数据库名称
  mysql_charset: "utf8mb4"      #编码格式
  mysql_parseTime: true         #是否解析时间格式(参考gorm文档)
  mysql_connPool: 10            #连接池数量
```

如何使用：
```go
var result map[string]interface{}
global.Service.Db.Raw("SELECT id, name, age FROM users WHERE name = ?", 3).Scan(&result)
global.Service.Log(core.InfoLevel, "result", result)
```
具体使用请参考gorm文档：https://gorm.io


### 十二、Redis缓存

框架集成了redis组件goredis，如需使用请在configs.yml增加如下配置：
```yml
service:
  ---

config:
  redis_open: false              #是否使用redis
  redis_addr: "127.0.0.1:6379"   #redis连接地址
  redis_pass: ""                 #redis密码
  redis_db: 0                    #数据库序号
```

如何使用：
```go
err := global.Service.Redis.Set(context.Background(), "key", "value", 0).Err()
if err != nil {
global.Service.Log(core.InfoLevel, "redis", err)
}
```
具体使用请参考goredis文档：https://github.com/go-redis/redis

### 十三、消息队列

框架集成rabbitmq（amqp协议都可）消息队列，如何使用：

```go
import (
    "github.com/panco95/go-garden/drives/amqp"
    "amqp2 "github.com/streadway/amqp"
)

// 连接
client, err := amqp.Conn("amqp://guest:guest@192.168.125.186:5672")
if err != nil {
    global.Service.Log(core.FatalLevel, "rabbitmq", err)
}
// 消费者
go func() {
    err := amqp.Consumer(client, "fanout", "test", "test", "test", func(msg amqp2.Delivery) {
        global.Service.Log(core.InfoLevel, "msg", msg.Body)
    }) 
	if err != nil {
        global.Service.Log(core.FatalLevel, "rabbitmq", err)
    }
}()
// 生产者
err = amqp.Publish(client, "fanout", "test", "test", "test", "test")
if err != nil {
    global.Service.Log(core.FatalLevel, "rabbitmq", err)
}
```

试着学习mysql与redis的自定义配置项，在你的项目里把rabbitmq连接地址从配置文件获取吧！

### 十四、负载均衡
上面的每一个服务都只启动了一个节点，同一份代码我们可以在多台服务器上启动，serviceName就是每个服务的标识，同名服务我们就称为服务集群；
复制一份user服务代码修改监听端口，启动；

现在user服务就是两个节点在运行，这时候我们调用user服务接口或者rpc方法的时候，go-garden内部会通过最小连接数以及轮询策略来选择服务器节点进行请求，开发者无需关心内部逻辑。

### 十五. 服务限流

在`config.yml`中我们可以给每个服务的每个接口配置单独的限流规则`limiter`参数，`5/1000`表示每5秒钟之内最多处理1000个请求，超出数量不会请求下游服务。

### 十六. 服务熔断

在`config.yml`中我们可以给每个服务的每个接口配置单独的熔断规则`fusing`参数，`5/100`表示接口每5秒钟之内下游服务器返回了100次错误响应后，直接会对下游服务熔断，在当前5秒内不请求下游服务，直接会返回错误响应。

### 十七. 服务重试

在调用下游服务时，下游服务可能会返回错误，go-garden支持重试机制，在config.yml中配置`callRetry`参数，格式 `timer1/timer2/timer3/...`，可不限制调整，重试次数使用`/`分隔，例如`100/200/200/200/500`表示重试5次，第一次100毫秒，第二次200毫秒，第三次200毫秒，第四次200毫秒，第五次500毫秒，如果重试第五次依然失败，会放弃重试返回错误。大家可根据项目自行调整重试策略配置。

### 十八. 超时控制

在调用下游服务时，下游服务可能会超时，go-garden支持超时控制防止超时问题加重导致服务雪崩，在routes.yml中给每个路由配置`timeout`参数，单位为毫秒ms，当下游服务接口请求超时将会熔断计数+1且不进行服务重试。

### 十九. 日志

提示：配置文件的`Debug`参数为`true`时，代表调试模式开启，任何日志输出都会同时打印在屏幕上和日志文件中，如果改为`false`，不会在屏幕打印，只会存储在日志文件中

日志文件路径：`runtime/logs`

go-garden封装了规范的日志函数，用如下代码进行调用：

```golang
    global.Service.Log(core.DebugLevel, "error", err)
    global.Service.Log(core.InfoLevel, "test", "info")
    global.Service.Log(core.WarnLevel, "test", "info")
    global.Service.Log(core.ErrorLevel, "test", "info")
    global.Service.Log(core.DPanicLevel, "test", "info")
    global.Service.Log(core.PanicLevel, "test", "info")
    global.Service.Log(core.FatalLevel, "test", "info")
```

第一个参数为日志级别，在源码`core/standard.go`文件中有定义，第二个参部为日志标识，第三个参数为日志内容，支持`error`或`string`。
