## go-garden命令行脚手架，快速创建项目

## 安装

`go install github.com/panco95/go-garden/tools/garden@v1.2.1`

## 创建项目

### 命令：

new {serviceName} {serviceType}

### 参数：

serviceName：服务名称

serviceType：服务类型，支持`gateway`和`service` (gateway表示网关，service表示业务)

### 示例

```shell
// 创建gateway网关类型服务 my-gateway
garden new my-gateway gateway
// 创建service业务类型服务 my-service
garden new my-service service
```

### 目录结构

|         目录/文件          |                              说明                               |
| ---------------------- | --------------------------------------------------------------- |
| api                           | api接口目录 |
| api/define.go                | 接口规范定义，包括状态码、提示信息、响应json格式                                                       |
| api/routes.go      | 当前服务接口路由定义，对应routes.yml中的path配置项                    |
| api/test.go      | api接口test业务逻辑                                                     |
| configs    |  配置文件目录                   |
| configs/config.yml    | 服务基础配置                                                     |
| configs/routes.yml  | 服务路由配置                            |
| global                   | 全局变量目录                           |
| global/global.go  |     全局变量定义文件                      |
| rpc    |  rpc定义目录                 |
| rpc/define   | 定义当前服务的rpc方法 调用参数与返回结果结构体                                                  |
| rpc/define/test.go   | 当前服务TestRpc方法参数、结果结构体
| rpc/base.go   | rpc基类
| rpc/test.go    |当前服务TestRpc方法具体逻辑
| runtime  | 服务启动后日志生成目录
| main.go  |服务驱动入口文件
| go.mod  | 依赖包管理文件
| auth/auth.go  | 网关统一鉴权中间件

### 特别注意

生成代码后需要修改配置文件方可启动服务，请查阅框架相关文档

### 使用答疑

脚手架创建项目后会自动安装依赖包`mod`，如果由于网络原因失败，请自行重试安装依赖包`mod`：

```
go mod tidy
```

