package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-ms/pkg"
	"go-ms/utils"
	"log"
	"os"
	"runtime"
	"time"
)

var (
	httpPort = flag.String("http_port", "8080", "Http listen port")
	rpcPort  = flag.String("rpc_port", "8180", "Rpc listen port")
	etcdAddr = flag.String("etcd_addr", "127.0.0.1:2379", "Etcd address, cluster format: 127.0.0.1:2379|127.0.0.1:2389")
	version  = flag.Bool("version", false, "Show version info")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("developing")
		os.Exit(0)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	var err error
	pkg.LogInit()
	err = pkg.EtcdRegister(*etcdAddr, *rpcPort, "gateway")
	if err != nil {
		log.Fatal("[Etcd register] ", err)
	}

	go httpServer(*httpPort)

	forever := make(chan bool)
	<-forever
}

func httpServer(port string) {
	gin.SetMode("release")
	server := gin.Default()
	path, _ := os.Getwd()
	err := utils.CreateDir(path + "/runtime")
	if err != nil {
		log.Fatal("[Create runtime folder] ", err)
	}
	file, err := os.Create(path + "/runtime/gin.log")
	if err != nil {
		log.Fatal("[Create gin log file] ", err)
	}
	gin.DefaultWriter = file
	server.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage)
	}))
	server.Use(gin.Recovery())
	err = server.Run(":" + port)
	if err != nil {
		log.Fatal("[Gin] ", err)
	}
}
