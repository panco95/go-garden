**goms：GO微服务框架**
<br>
<br>
**说明：**<br>
本项目是由个人开发的微服务基础框架，项目正在积极开发中，很期待得到你的star。<br>

**已实现功能：**<br>
1、服务注册发现<br>
2、网关路由分发<br>
3、负载均衡<br>
4、服务调用安全验证<br>
5、服务重试<br>
6、分布式链路追踪日志<br>

**准备工作：**<br>
1、Etcd，Docker启动：<br>
```
docker run --rm -it -d --name etcd -p 2379:2379 -e "ALLOW_NONE_AUTHENTICATION=yes" -e "ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379" bitnami/etcd
```
<br>
2、Elasticsearch，Docker启动：<br>

```
docker network create es
docker run --rm -it -d --name elasticsearch --net es -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" -e "cluster.name=goms" elasticsearch:7.14.1
docker run --rm -it -d --name kibana --net es -p 5601:5601 -e "discovery.type=single-node" kibana:7.14.1
```
<br>
3、Rabbitmq，Docker启动：<br>

```
docker run --rm -it -d --name rabbitmq --hostname rabbitmq  -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

**文档**<br>

**一、配置文件**<br><br>
路径：config/config.yml<br>
callServiceKey: 服务调用安全验证key<br>
etcdAddr: etcd服务地址，集群可多行填写<br>
esAddr: elasticsearch服务地址<br>
amqpAddr：amqp消息队列服务地址，本项目使用的是rabbitmq<br>
services：服务配置<br>
配置示例：

```
callServiceKey: "goms by panco"
etcdAddr:
  - "192.168.125.181:2379"
esAddr: "http://192.168.125.181:9200"
amqpaddr: "amqp://guest:guest@192.168.125.181:5672"
services:
  user:
    register: "/register"
    login: "/login"
  order:
    submit: "/order/submit"
```

<br>


**二、Gateway API网关**<br>

启动：`go run example/gateway/main.go`<br>
参数：<br>
-http_port：http监听端口<br>
-rpc_port：rpc监听端口<br>

模拟三个Gateway网关节点进行集群架设：<br>
```
go run example/gateway/main.go -http_port 8080 -rpc_port 8180
go run example/gateway/main.go -http_port 8081 -rpc_port 8181
go run example/gateway/main.go -http_port 8082 -rpc_port 8182
```

观察我们第一个gateway节点打印如下：<br>
```
2021/08/10 11:08:01 [gateway] Http Listen on port: 8080
2021/08/10 11:08:14 [gateway] node [192.168.125.179:8181_192.168.125.179:8081] join
2021/08/10 11:08:23 [gateway] node [192.168.125.179:8182_192.168.125.179:8082] join
```

表示gateway节点互相能够发现其他节点服务启动

<br>

**三、Service服务**<br>

启动服务节点：<br>
`go run example/services/user/main.go`<br>
参数：<br>
-rpc_port：rpc监听端口<br>
-etcd_addr：etcd服务地址（支持集群格式：`127.0.0.1:2379|127.0.0.1:2380`）<br>

模拟三个user服务节点进行集群架设：<br>
```
go run example/services/user/main.go -http_port 9080 -rpc_port 9180
go run example/services/user/main.go -http_port 9081 -rpc_port 9181
go run example/services/user/main.go -http_port 9082 -rpc_port 9182
```

观察我们第一个node节点打印如下：<br>
```
2021/08/10 11:10:46 [user] Http Listen on port: 9080
2021/08/10 11:10:54 [user] node [192.168.125.179:9181_192.168.125.179:9081] join
2021/08/10 11:11:02 [user] node [192.168.125.179:9182_192.168.125.179:9082] join
```
<br>

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
`etcdctl get --prefix goms`<br>
（小提示：Docker启动的etcd需要进入etcd容器查看）<br>
`docker container exec -ti [容器id] bash`<br>
打印如下：<br>

```
goms_gateway_192.168.125.179:8180_192.168.125.179:8080
0
goms_gateway_192.168.125.179:8181_192.168.125.179:8081
0
goms_gateway_192.168.125.179:8182_192.168.125.179:8082
0
goms_user_192.168.125.179:9181_192.168.125.179:9081
0
goms_user_192.168.125.179:9182_192.168.125.179:9082
0
```

<br>
可以看出，我们刚刚启动的所以服务节点都在etcd里面以goms_前缀存储，后面跟着服务名称，然后是两个服务地址，前面的是RPC监听地址，后面的是HTTP监听地址<br><br>

**四、开启trace链路跟踪日志消费服务，可以多开增加消费并发速度：**<br>
```
go run example\traceConsumer\main.go
```
<br>

**五、访问API网关**<br>

1、集群服务状态查看：<br>
打开浏览器，输入gateway任一节点http地址，后面加上/cluster：
`http://127.0.0.1:8082/cluster` <br>

```
{
    "services": {
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
            "PollNext": 0,
            "Nodes": [
                "192.168.125.179:9180_192.168.125.179:9080",
                "192.168.125.179:9181_192.168.125.179:9081",
                "192.168.125.179:9182_192.168.125.179:9082"
            ],
            "RequestFinish": 0
        }
    },
    "status": true
}
```

可以看到当前整个微服务的所有服务所有节点的信息，PollNext是负载均衡轮询策略索引，Nodes就是当前在线服务集群列表，RequestFinish是当前服务集群一共处理的请求数。<br>

2、打开postman，选择GET/POST请求方式，URL输入gateway任一节点http地址，然后加上访问用户服务的路由(config/services.yml)如：
``http://192.168.125.179:8081/api/user/login?a=1&b=3
``，可增加任意header和url_param，body支持application/json、form-data、application/x-www-form-urlencoded，最后gateway都会封装为application/json请求下游服务，请求返回json数据：<br>

```
{
    "code": 0,
    "data": {
        "body": {
            "password": "test",
            "username": "test"
        },
        "headers": {
            "Accept": "*/*",
            "Accept-Encoding": "gzip, deflate, br",
            "Connection": "keep-alive",
            "Content-Length": "37",
            "Content-Type": "application/json",
            "Postman-Token": "85d0aef3-4f06-4887-9ec8-0531c09d209c",
            "User-Agent": "PostmanRuntime/7.28.4"
        },
        "method": "POST",
        "urlParam": "?a=1&b=3"
    },
    "msg": "success",
    "status": true
}
```

3、再看集群服务状态

```
{
    "services": {
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
    },
    "status": true
}
```

可以看到user集群的的PollNext已经变成1，表示下一个请求将由第二个Node来处理请求。多请求几次看看！切换user服务的控制台显示日志打印信息，会看到三个user服务依次轮流收到请求<br><br>

**六、查看trace链路跟踪日志：**<br>
1、切换到trace消费者命令行窗口，查看消费成功打印：
```
2021/09/04 17:43:07 [amqp] trace consumer is runner
2021/09/04 17:43:07 amqp trace consumer success
2021/09/04 17:43:07 amqp trace consumer success
2021/09/04 17:43:07 amqp trace consumer success
2021/09/04 17:43:07 amqp trace consumer success
```
2、浏览器登录kibana `http://127.0.0.1:5601`<br>
3、点击dev tools<br>
4、查询日志数据：<br>

```
GET /trace_logs/_search
{
  "query": {
    "match_all": {}
  },
  "from":0,
  "size":1000
}
```

链路跟踪日志，一个客户端请求会有多条日志，按照 requestId进行归类是否同一请求产生的日志：
```
{
  "took" : 1,
  "timed_out" : false,
  "_shards" : {
    "total" : 1,
    "successful" : 1,
    "skipped" : 0,
    "failed" : 0
  },
  "hits" : {
    "total" : {
      "value" : 8,
      "relation" : "eq"
    },
    "max_score" : 1.0,
    "hits" : [
      {
        "_index" : "trace_logs",
        "_type" : "_doc",
        "_id" : "USQJsHsBU4boCFpVSHWL",
        "_score" : 1.0,
        "_source" : {
          "requestId" : "2cf46aa2-519d-4517-8465-6efae813936d",
          "request" : {
            "method" : "POST",
            "url" : "/api/user/login",
            "urlParam" : "?field_1=1&field_2=2=",
            "clientIp" : "127.0.0.1",
            "headers" : {
              "Accept" : "*/*",
              "Accept-Encoding" : "gzip, deflate, br",
              "Call-Service-Key" : "go-ms by panco",
              "Connection" : "keep-alive",
              "Content-Length" : "265",
              "Content-Type" : "multipart/form-data; boundary=--------------------------568994750929172835812890",
              "Postman-Token" : "6ef9df7e-4f2e-445c-bec7-636fc6fe7941",
              "User-Agent" : "PostmanRuntime/7.28.4"
            },
            "body" : {
              "ds1" : "11",
              "ds2" : "111"
            }
          },
          "event" : "request.start",
          "time" : "2021-09-04 17:00:03",
          "serviceName" : "gateway",
          "serviceId" : "goms_gateway_192.168.8.98:8180_192.168.8.98:8080",
          "projectName" : "goms",
          "trace" : null
        }
      },
      {
        "_index" : "trace_logs",
        "_type" : "_doc",
        "_id" : "UiQJsHsBU4boCFpVSHXJ",
        "_score" : 1.0,
        "_source" : {
          "requestId" : "2cf46aa2-519d-4517-8465-6efae813936d",
          "request" : {
            "method" : "POST",
            "url" : "/login",
            "urlParam" : "?field_1=1&field_2=2=",
            "clientIp" : "192.168.8.98",
            "headers" : {
              "Accept" : "*/*",
              "Accept-Encoding" : "gzip, deflate, br",
              "Call-Service-Key" : "goms by panco",
              "Connection" : "keep-alive",
              "Content-Length" : "14",
              "Content-Type" : "application/x-www-form-urlencoded",
              "Postman-Token" : "6ef9df7e-4f2e-445c-bec7-636fc6fe7941",
              "User-Agent" : "PostmanRuntime/7.28.4",
              "X-Request-Id" : "2cf46aa2-519d-4517-8465-6efae813936d"
            },
            "body" : {
              "ds1" : "11",
              "ds2" : "111"
            }
          },
          "event" : "service.start",
          "time" : "2021-09-04 17:00:03",
          "serviceName" : "user",
          "serviceId" : "goms_user_192.168.8.98:9180_192.168.8.98:9080",
          "projectName" : "goms",
          "trace" : null
        }
      },
      {
        "_index" : "trace_logs",
        "_type" : "_doc",
        "_id" : "UyQJsHsBU4boCFpVSHX9",
        "_score" : 1.0,
        "_source" : {
          "requestId" : "2cf46aa2-519d-4517-8465-6efae813936d",
          "request" : {
            "method" : "POST",
            "url" : "/login",
            "urlParam" : "?field_1=1&field_2=2=",
            "clientIp" : "192.168.8.98",
            "headers" : {
              "Accept" : "*/*",
              "Accept-Encoding" : "gzip, deflate, br",
              "Call-Service-Key" : "goms by panco",
              "Connection" : "keep-alive",
              "Content-Length" : "14",
              "Content-Type" : "application/x-www-form-urlencoded",
              "Postman-Token" : "6ef9df7e-4f2e-445c-bec7-636fc6fe7941",
              "User-Agent" : "PostmanRuntime/7.28.4",
              "X-Request-Id" : "2cf46aa2-519d-4517-8465-6efae813936d"
            },
            "body" : {
              "ds1" : "11",
              "ds2" : "111"
            }
          },
          "event" : "service.end",
          "time" : "2021-09-04 17:00:03",
          "serviceName" : "user",
          "serviceId" : "goms_user_192.168.8.98:9180_192.168.8.98:9080",
          "projectName" : "goms",
          "trace" : {
            "timing" : "17.9651ms"
          }
        }
      },
      {
        "_index" : "trace_logs",
        "_type" : "_doc",
        "_id" : "VCQJsHsBU4boCFpVSXUq",
        "_score" : 1.0,
        "_source" : {
          "requestId" : "2cf46aa2-519d-4517-8465-6efae813936d",
          "request" : {
            "method" : "POST",
            "url" : "/api/user/login",
            "urlParam" : "?field_1=1&field_2=2=",
            "clientIp" : "127.0.0.1",
            "headers" : {
              "Accept" : "*/*",
              "Accept-Encoding" : "gzip, deflate, br",
              "Call-Service-Key" : "go-ms by panco",
              "Connection" : "keep-alive",
              "Content-Length" : "265",
              "Content-Type" : "multipart/form-data; boundary=--------------------------568994750929172835812890",
              "Postman-Token" : "6ef9df7e-4f2e-445c-bec7-636fc6fe7941",
              "User-Agent" : "PostmanRuntime/7.28.4"
            },
            "body" : {
              "ds1" : "11",
              "ds2" : "111"
            }
          },
          "event" : "request.end",
          "time" : "2021-09-04 17:00:03",
          "serviceName" : "gateway",
          "serviceId" : "goms_gateway_192.168.8.98:8180_192.168.8.98:8080",
          "projectName" : "goms",
          "trace" : {
            "timing" : "48.2256ms"
          }
        }
      },
      {
        "_index" : "trace_logs",
        "_type" : "_doc",
        "_id" : "VSQNsHsBU4boCFpVw3XF",
        "_score" : 1.0,
        "_source" : {
          "requestId" : "09978d95-2c8f-4088-9cb8-f324dc9dec5e",
          "request" : {
            "method" : "POST",
            "url" : "/api/user/login",
            "urlParam" : "?field_1=1&field_2=2=",
            "clientIp" : "127.0.0.1",
            "headers" : {
              "Accept" : "*/*",
              "Accept-Encoding" : "gzip, deflate, br",
              "Call-Service-Key" : "go-ms by panco",
              "Connection" : "keep-alive",
              "Content-Length" : "265",
              "Content-Type" : "multipart/form-data; boundary=--------------------------860243771345130791268878",
              "Postman-Token" : "965cf7a2-1637-4b03-83ab-df123614ebea",
              "User-Agent" : "PostmanRuntime/7.28.4"
            },
            "body" : {
              "ds1" : "11",
              "ds2" : "111"
            }
          },
          "event" : "request.start",
          "time" : "2021-09-04 17:04:57",
          "serviceName" : "gateway",
          "serviceId" : "goms_gateway_192.168.8.98:8180_192.168.8.98:8080",
          "projectName" : "goms",
          "trace" : null
        }
      },
      {
        "_index" : "trace_logs",
        "_type" : "_doc",
        "_id" : "ViQNsHsBU4boCFpVw3Xb",
        "_score" : 1.0,
        "_source" : {
          "requestId" : "09978d95-2c8f-4088-9cb8-f324dc9dec5e",
          "request" : {
            "method" : "POST",
            "url" : "/login",
            "urlParam" : "?field_1=1&field_2=2=",
            "clientIp" : "192.168.8.98",
            "headers" : {
              "Accept" : "*/*",
              "Accept-Encoding" : "gzip, deflate, br",
              "Call-Service-Key" : "goms by panco",
              "Connection" : "keep-alive",
              "Content-Length" : "14",
              "Content-Type" : "application/x-www-form-urlencoded",
              "Postman-Token" : "965cf7a2-1637-4b03-83ab-df123614ebea",
              "User-Agent" : "PostmanRuntime/7.28.4",
              "X-Request-Id" : "09978d95-2c8f-4088-9cb8-f324dc9dec5e"
            },
            "body" : {
              "ds1" : "11",
              "ds2" : "111"
            }
          },
          "event" : "service.start",
          "time" : "2021-09-04 17:04:57",
          "serviceName" : "user",
          "serviceId" : "goms_user_192.168.8.98:9180_192.168.8.98:9080",
          "projectName" : "goms",
          "trace" : null
        }
      },
      {
        "_index" : "trace_logs",
        "_type" : "_doc",
        "_id" : "VyQNsHsBU4boCFpVw3X4",
        "_score" : 1.0,
        "_source" : {
          "requestId" : "09978d95-2c8f-4088-9cb8-f324dc9dec5e",
          "request" : {
            "method" : "POST",
            "url" : "/login",
            "urlParam" : "?field_1=1&field_2=2=",
            "clientIp" : "192.168.8.98",
            "headers" : {
              "Accept" : "*/*",
              "Accept-Encoding" : "gzip, deflate, br",
              "Call-Service-Key" : "goms by panco",
              "Connection" : "keep-alive",
              "Content-Length" : "14",
              "Content-Type" : "application/x-www-form-urlencoded",
              "Postman-Token" : "965cf7a2-1637-4b03-83ab-df123614ebea",
              "User-Agent" : "PostmanRuntime/7.28.4",
              "X-Request-Id" : "09978d95-2c8f-4088-9cb8-f324dc9dec5e"
            },
            "body" : {
              "ds1" : "11",
              "ds2" : "111"
            }
          },
          "event" : "service.end",
          "time" : "2021-09-04 17:04:57",
          "serviceName" : "user",
          "serviceId" : "goms_user_192.168.8.98:9180_192.168.8.98:9080",
          "projectName" : "goms",
          "trace" : {
            "timing" : "13.2229ms"
          }
        }
      },
      {
        "_index" : "trace_logs",
        "_type" : "_doc",
        "_id" : "WCQNsHsBU4boCFpVxHUF",
        "_score" : 1.0,
        "_source" : {
          "requestId" : "09978d95-2c8f-4088-9cb8-f324dc9dec5e",
          "request" : {
            "method" : "POST",
            "url" : "/api/user/login",
            "urlParam" : "?field_1=1&field_2=2=",
            "clientIp" : "127.0.0.1",
            "headers" : {
              "Accept" : "*/*",
              "Accept-Encoding" : "gzip, deflate, br",
              "Call-Service-Key" : "go-ms by panco",
              "Connection" : "keep-alive",
              "Content-Length" : "265",
              "Content-Type" : "multipart/form-data; boundary=--------------------------860243771345130791268878",
              "Postman-Token" : "965cf7a2-1637-4b03-83ab-df123614ebea",
              "User-Agent" : "PostmanRuntime/7.28.4"
            },
            "body" : {
              "ds1" : "11",
              "ds2" : "111"
            }
          },
          "event" : "request.end",
          "time" : "2021-09-04 17:04:57",
          "serviceName" : "gateway",
          "serviceId" : "goms_gateway_192.168.8.98:8180_192.168.8.98:8080",
          "projectName" : "goms",
          "trace" : {
            "timing" : "60.9127ms"
          }
        }
      }
    ]
  }
}
```