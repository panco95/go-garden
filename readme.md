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

## 教程：基于Go Garden快速构建微服务
访问 [examples](https://github.com/panco95/go-garden/tree/master/examples) 查看教程

## 示例
访问 [examples](https://github.com/panco95/go-garden/tree/master/docs/tutorial.md) 查看详细使用示例

## 许可证

Go Garden 包含 Apache 2.0 许可证