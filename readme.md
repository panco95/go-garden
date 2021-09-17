# Go Garden [![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Go Garden是一款面向分布式系统架构的微服务框架

## 概念

Go Garden为分布式系统架构的开发提供了核心需求，包括微服务的基础架构支持，例如gateway网关模块做路由分发支持，服务调用链路追踪的集成。

## 特性

- **服务注册发现**

- **网关路由分发**

- **负载均衡**

- **动态配置**

- **配置同步**

- **安全认证**

- **服务重试机制**

- **分布式链路追踪**

- **常用可选组件**

## 开发者必读

Go Garden考虑到开发者的使用门槛，并没有实现Http框架，而是集成了评价较高的Gin框架。这样可以大大减少开发者的学习路径，如果你是Go初学者，请提前阅读Gin框架文档。


## 快速开始

```golang
import "github.com/panco95/go-garden"

// initialise
garden.Init()
// start the service
garden.Run(Route, Auth)
```

## 教程：基于Go Garden快速构建微服务
访问 [基于Go Garden快速构建微服务](docs/tutorial.md) 学习如何使用Go
Garden

## 示例代码
访问 [examples](examples) 查看教程

## 许可证

Go Garden 包含 Apache 2.0 许可证