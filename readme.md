# Go Garden 
[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/panco95/go-garden) [![Go Report Card](https://goreportcard.com/badge/github.com/panco95/go-garden)](https://goreportcard.com/report/github.com/panco95/go-garden) 


go-garden是一款面向分布式系统架构的分布式服务框架

## 概念

* 框架为分布式系统架构的开发提供了核心需求，包括微服务的一些基础架构支持，减少开发者对微服务的基础开发，更着力于业务开发；
* 框架支持Http/Rpc协议，api接口使用http协议，rpc方法使用rpc协议，统一处理负载均衡、限流、熔断、超时等策略；
* 框架没有集成数据库、缓存之类的扩展，这里考虑到使用者对服务的设计可能会使用到不同的包，建议开发者自己导入这类扩展包使用；
* 框架不限制代码结构，只需要配置文件和几行代码就可以启动一个服务，项目的结构完全由开发者自行设计，建议大家使用脚手架工具生成项目结构。

## 特性

- **服务注册发现**

- **网关路由分发**

- **网关负载均衡**

- **Rpc/Http协议**

- **可配服务限流**

- **可配服务熔断**

- **可配服务重试**

- **可配超时控制**

- **动态路由配置**

- **集群自动同步**

- **调用安全认证**

- **分布式链路追踪**

- **统一日志存储**

- **脚手架工具**


## 快速开始

```
// 安装项目脚手架
go install github.com/panco95/go-garden/tools/garden@v1.1.4

// 创建项目
garden new my-gateway gateway
garden new my-service service

// 修改服务配置和路由配置
......

// 启动网关
go run my-gateway/main.go
// 启动服务
go run my-service/main.go
```

## 教程：基于Go Garden快速构建微服务
访问 [基于Go Garden快速构建微服务](docs/tutorial.md) 跟着一步一步学习如何使用go-garden

## 教程：代码示例
访问 [examples](examples) 查看完整示例项目

## 脚手架：快速创建按项目
访问 [tools](tools/garden) 查看脚手架使用说明

## 特别鸣谢
感谢 **JetBrains** 为本项目免费提供的正版 **IDE** 激活码，支持正版请前往购买：[Jetbrains商店](https://www.jetbrains.com/store/#commercial?billing=yearly)

## 联系我们
作者微信：freePan_1995

## 许可证
Go Garden 包含 Apache 2.0 许可证
