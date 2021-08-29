package base

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-ms/utils"
	"log"
	"os"
	"time"
)

func HttpServer(port, serverName string, route func(r *gin.Engine)) {
	gin.SetMode("release")
	server := gin.Default()
	path, _ := os.Getwd()
	err := utils.CreateDir(path + "/runtime")
	if err != nil {
		log.Fatal("[Create runtime folder] ", err)
	}
	file, err := os.Create(fmt.Sprintf("%s/runtime/gin_%s.log", path, serverName))
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
	route(server)

	log.Printf("[%s] Http Listen on port: %s", serverName, port)
	log.Fatal(server.Run(":" + port))
}

func RpcServer(port, serverName string) {
	//listenAddress := "127.0.0.1:" + port
	//listen, err := net.Listen("tcp", listenAddress)
	//if err != nil {
	//	log.Fatal("[RPC] " + err.Error())
	//}
	//
	//s := grpc.NewServer()
	//var Service = client.UserService{}
	//pb.RegisterUserServer(s, Service)
	//
	//log.Printf("[RPC][%s service] Listen on port: %s", serverName, port)
	//log.Fatal(s.Serve(listen))
}
