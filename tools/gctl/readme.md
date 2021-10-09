## go-garden命令行工具gctl，快速创建项目

### 1、安装

`go install github.com/panco95/go-garden/tools/gctl@latest`

### 2. 参数

* -name：服务名称
* -class：服务类型，支持`gateway`和`service` (gateway表示网关服务，service表示业务服务)

### 3. 示例（创建三个demo项目）
```shell
gctl -name demo-gateway -class gateway
gctl -name demo-user -class service
gctl -name demo-pay -class service
```

### 4、使用答疑

`gctl`会自动创建服务文件夹，以及自动安装`mod`依赖，如果由于网络原因失败，请自行修改`proxy`并且进入服务文件夹执行：

```
go mod tidy
```

