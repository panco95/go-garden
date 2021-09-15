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

- **可选组件：消息队列、Redis、ES**

- **HTTP服务基于GIN开发，适合大部分GO开发者，简单易学轻量**


## 快速开始

快速使用Go Garden

```golang
import "github.com/panco95/go-garden"

// initialise
garden.Init()
// start the service
garden.Run(Route, Auth)
```

访问 [examples](https://github.com/panco95/go-garden/tree/master/examples) 查看详细使用示例，包括gateway(api网关)、user(用户中心)、pay(支付中心)示例

## 许可证

Go Garden 包含 Apache 2.0 许可证