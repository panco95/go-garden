## go-garden命令行脚手架，快速创建项目

## 安装

`go install github.com/panco95/go-garden/tools/garden@v1.1.11`


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

### 特别注意
生成代码后需要修改配置文件方可启动服务，请查阅框架相关文档

### 使用答疑

脚手架创建项目后会自动安装依赖包`mod`，如果由于网络原因失败，请自行重试安装依赖包`mod`：

```
go mod tidy
```

