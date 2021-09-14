**go-garden：GO微服务框架**
<br>
<br>
**说明：**<br>
go-garden是由个人开发的微服务基础框架，项目正在积极开发中，很期待得到你的star。<br>

**特性：**<br>
1、服务注册发现<br>
2、网关路由分发<br>
3、负载均衡策略<br>
4、服务调用安全<br>
5、服务重试策略<br>
6、分布式链路追踪<br>
7、可选组件Redis、Rabbitmq、Elasticsearch<br>
8、支持Gin框架进行开发<br>

**准备工作：**<br>
1、Etcd，Docker启动：<br>
```
docker run --rm -it -d --name etcd -p 2379:2379 -e "ALLOW_NONE_AUTHENTICATION=yes" -e "ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379" bitnami/etcd
```
<br>
2、Zipkin，Docker启动：<br>

```
docker run --rm -it -d --name zipkin -p 9411:9411 openzipkin/zipkin
```

3、（示例程序）依赖redis数据库，Docker启动：<br>
```
docker run --rm -it -d --name redis -p 6379:6379 redis
```

<br>

一、配置文件说明：<br>
1、configs\config.yml<br>
必填配置项：
ProjectName：项目名称<br>
ServiceName：服务名称<br>
HttpPort：http监听端口<br>
RpcPort：rpc监听端口<br>
CallServiceKey: 服务调用安全验证key<br>
EtcdAddress: etcd地址，集群可多行填写<br>
ZipkinAddress: zipkin地址<br><br>
可选配置项，不需要可注释掉：<br>
RedisAddress: redis地址<br>
ElasticsearchAddress: elasticsearch地址<br>
AmqpAddress: rabbitmq地址<br>

2、configs\services.yml<br>
这是服务调用路由配置，格式：

```
services:
  service1:
    action1: "/route1"
    action2: "/route2"
  service2:
    action1: "/route1"
```
说明：访问网关 /api/serviceName/actionName 会自动转发到服务接口 /route1 路由

