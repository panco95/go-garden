package core

import "github.com/gin-gonic/gin"

// Run  http(gin) && rpc(rpcx)
func (g *Garden) Run(route func(r *gin.Engine), rpc interface{}, auth func() gin.HandlerFunc) {
	go g.runHttpServer(route, auth)
	go g.runRpcServer(rpc)
	forever := make(chan int, 0)
	<-forever
}

func (g *Garden) runHttpServer(route func(r *gin.Engine), auth func() gin.HandlerFunc) {
	address := g.ServiceIp
	if g.cfg.Service.HttpOut {
		address = "0.0.0.0"
	}
	listenAddress := address + ":" + g.cfg.Service.HttpPort
	if err := g.ginListen(listenAddress, route, auth); err != nil {
		g.Log(FatalLevel, "ginRun", err)
	}
}

func (g *Garden) runRpcServer(rpc interface{}) {
	address := g.ServiceIp
	if g.cfg.Service.RpcOut {
		address = "0.0.0.0"
	}
	rpcAddress := address + ":" + g.cfg.Service.RpcPort
	if err := g.rpcListen(g.cfg.Service.ServiceName, "tcp", rpcAddress, rpc, ""); err != nil {
		g.Log(FatalLevel, "rpcRun", err)
	}
}

// RebootFunc if func panic
func (g *Garden) RebootFunc(label string, f func()) {
	defer func() {
		if err := recover(); err != nil {
			g.Log(ErrorLevel, label, err)
			f()
		}
	}()
	f()
}
