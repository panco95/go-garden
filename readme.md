# Go Garden 
[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/panco95/go-garden) [![Go Report Card](https://goreportcard.com/badge/github.com/panco95/go-garden)](https://goreportcard.com/report/github.com/panco95/go-garden) 


go-garden是一款面向分布式系统架构的分布式服务框架

github地址：https://github.com/panco95/go-garden

码云地址：https://gitee.com/pancoJ/go-garden

## 概念

* 为分布式系统架构的开发提供了核心需求，包括微服务的一些基础架构支持，减少开发者对微服务的基础开发，更着力于业务开发；
* 支持Http/Rpc协议，http使用gin，rpc使用rpcx；
* RPC无需定义Proto文件，只需要定义结构体即可；
* 使用脚手架快速创建项目。

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

- **安全调用认证**

- **分布式链路追踪**

- **服务监控与警报**

- **统一日志存储**

- **脚手架工具**


## 快速开始

```
// 安装项目脚手架
go get -u github.com/panco95/go-garden@v1.4.4
go install github.com/panco95/go-garden/tools/garden@v1.4.4

// 创建项目
garden new my-gateway gateway
garden new my-service service

// 修改服务配置和路由配置
......

// 启动网关
go run my-gateway/main.go -configs=my-gateway/configs -runtime=my-gateway/runtime
// 启动服务
go run my-service/main.go -configs=my-gateway/configs -runtime=my-gateway/runtime
```

## 教程：基于Go Garden快速构建微服务
访问 [基于Go Garden快速构建微服务](docs/tutorial.md) 跟着一步一步学习如何使用go-garden

## 脚手架：快速创建项目
访问 [tools](tools/garden) 查看脚手架使用说明

## Docker部署：
参考deply目录下的Dockerfile以及docker-compose.yml文件
## 交流
QQ群：967256601

## 许可证
Go Garden 包含 Apache 2.0 许可证
