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

- **HTTP服务基于GIN开发，适合大部分GO开发者，简单易学轻量**


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

* 创建quick_gateway目录后进入目录
* 执行 `go mod init` 初始化一个go项目
* 新建go程序入口文件 `main.go` 并输入以下代码：
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
* 安装go mod包： `go mod tidy`
* 执行程序：`go run main.go`

## 许可证

Go Garden 包含 Apache 2.0 许可证