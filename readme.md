# Go Garden 
[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/panco95/go-garden) [![Go Report Card](https://goreportcard.com/badge/github.com/panco95/go-garden)](https://goreportcard.com/report/github.com/panco95/go-garden) 


go-garden是一款面向分布式系统架构的微服务框架

## 概念

* go-garden为分布式系统架构的开发提供了核心需求，包括微服务的一些基础架构支持，减少开发者对微服务的基础开发，更着力于业务开发；
* go-garden支持Http/Rpc协议，http底层框架使用Gin，开发者如不熟悉Gin，可先去看看Gin文档方便更好的使用go-garden；
* go-garden并没有集成数据库、缓存之类的扩展，这里考虑到使用者对服务的设计可能会使用到不同的包，建议开发者自己导入这类扩展包使用；
* go-garden不限制代码结构，只需要配置文件和几行代码就可以启动一个服务，项目的结构完全由开发者自行设计，example示例代码中使用的项目结构可供大家参考。

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

- **日志系统**

- **项目脚手架**

## 代码预览

```golang
import "github.com/panco95/go-garden/core"

var service *core.Garden

func main() {
    service = core.New()
    service.Run(api.Routes, new(rpc.Rpc), auth.Auth)
}
```

## 教程：基于Go Garden快速构建微服务
访问 [基于Go Garden快速构建微服务](docs/tutorial.md) 跟着一步一步学习如何使用go-garden

## 教程：代码示例
访问 [examples](examples) 查看示例教程代码

## 特别鸣谢
感谢 **JetBrains** 为本项目免费提供的正版 **IDE** 激活码，支持正版请前往购买：[Jetbrains商店](https://www.jetbrains.com/store/#commercial?billing=yearly)

## 联系我们
作者微信：freePan_1995

## 许可证
Go Garden 包含 Apache 2.0 许可证
