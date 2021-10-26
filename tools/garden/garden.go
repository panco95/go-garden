package main

import (
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	var cmdNew = &cobra.Command{
		Use:   "new [serviceName] [serviceType]",
		Short: "Create new service project.",
		Long:  `serviceName: name | serviceType: gateway or service`,
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			serviceName := args[0]
			serviceType := args[1]
			if serviceName == "" {
				log.Fatal("Empty serviceName!")
			}
			if serviceType != "gateway" && serviceType != "service" {
				log.Fatal("ServiceType just support: gateway or service!")
			}
			switch serviceType {
			case "gateway":
				newGateway(args[0])
				break
			case "service":
				newService(args[0])
				break
			}
		},
	}

	var rootCmd = &cobra.Command{Use: "garden"}
	rootCmd.AddCommand(cmdNew)
	rootCmd.Execute()
}

func newGateway(serviceName string) {
	createDir("./" + serviceName)
	createFile("./"+serviceName+"/main.go", gatewayMain(serviceName))
	createDir("./" + serviceName + "/auth")
	createFile("./"+serviceName+"/auth/auth.go", gatewayAuth())
	createDir("./" + serviceName + "/configs")
	createFile("./"+serviceName+"/configs/routes.yml", gatewayRoutesYml())
	createFile("./"+serviceName+"/configs/config.yml", configsYml(serviceName))
	createDir("./" + serviceName + "/global")
	createFile("./"+serviceName+"/global/global.go", globalGo())
	createDir("./" + serviceName + "/rpc")
	createFile("./"+serviceName+"/rpc/base.go", gatewayRpcBase())
	sysCmd(serviceName, "go", "mod", "init", serviceName)
	sysCmd(serviceName, "go", "mod", "tidy")
}

func newService(serviceName string) {
	createDir("./" + serviceName)
	createFile("./"+serviceName+"/main.go", serviceMain(serviceName))
	createDir("./" + serviceName + "/configs")
	createFile("./"+serviceName+"/configs/routes.yml", serviceRoutesYml(serviceName))
	createFile("./"+serviceName+"/configs/config.yml", configsYml(serviceName))
	createDir("./" + serviceName + "/global")
	createFile("./"+serviceName+"/global/global.go", globalGo())
	createDir("./" + serviceName + "/rpc")
	createFile("./"+serviceName+"/rpc/base.go", rpcBase())
	createFile("./"+serviceName+"/rpc/test.go", rpcTest(serviceName))
	createDir("./" + serviceName + "/rpc/define")
	createFile("./"+serviceName+"/rpc/define/test.go", rpcDefineTest())
	createDir("./" + serviceName + "/api")
	createFile("./"+serviceName+"/api/base.go", apiBase(serviceName))
	createFile("./"+serviceName+"/api/test.go", apiTest(serviceName))
	sysCmd(serviceName, "go", "mod", "init", serviceName)
	sysCmd(serviceName, "go", "mod", "tidy")
}

func createFile(path string, data string) {
	if err := ioutil.WriteFile(path, []byte(data), 0777); err != nil {
		log.Fatal(err)
	}
}

func createDir(path string) {
	exists, err := pathExists(path)
	if err != nil {
		log.Fatal(err)
	}
	if exists {
		log.Fatal("Dir is exists!")
	}
	if err := os.Mkdir(path, os.ModePerm); err != nil {
		log.Print(err)
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func sysCmd(name string, command string, param ...string) {
	cmd := exec.Command(command, param...)
	cmd.Dir = "./" + name
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	if _, err := ioutil.ReadAll(stdout); err != nil {
		log.Fatal(err)
	}
}

func gatewayMain(serviceName string) string {
	return strings.Replace("package main\n\nimport (\n\t\"github.com/panco95/go-garden/core\"\n\t\"<>/auth\"\n\t\"<>/global\"\n\t\"<>/rpc\"\n)\n\nfunc main() {\n\tglobal.Service = core.New()\n\tglobal.Service.Run(global.Service.GatewayRoute, new(rpc.Rpc), auth.Auth)\n}\n", "<>", serviceName, 999)
}

func gatewayAuth() string {
	return "package auth\n\nimport \"github.com/gin-gonic/gin\"\n\n// Auth Customize the auth middleware\nfunc Auth() gin.HandlerFunc {\n\treturn func(c *gin.Context) {\n\t\t// before logic\n\t\tc.Next()\n\t\t// after logic\n\t}\n}\n"
}

func configsYml(serviceName string) string {
	return strings.Replace("service:\n  debug: true\n  serviceName: <>\n  httpOut: true\n  httpPort: 8080\n  rpcOut: false\n  rpcPort: 9000\n  callKey: garden\n  callRetry: 20/30/50\n  etcdKey: garden\n  etcdAddress:\n    - 127.0.0.1:2379\n  zipkinAddress: http:/127.0.0.1:9411/api/v2/spans\n\nconfig:\n", "<>", serviceName, 999)
}

func serviceRoutesYml(serviceName string) string {
	return strings.Replace("routes:\n  <>:\n    test:\n      type: http\n      path: /test\n      limiter: 5/100\n      fusing: 5/100\n      timeout: 2000\n    TestRpc:\n      type: rpc\n      limiter: 5/100\n      fusing: 5/100\n      timeout: 2000", "<>", serviceName, 999)
}

func gatewayRoutesYml() string {
	return "routes:\n r <xxx>:\n    test:\n      type: http\n      path: /test\n      limiter: 5/100\n      fusing: 5/100\n      timeout: 2000\n    TestRpc:\n      type: rpc\n      limiter: 5/100\n      fusing: 5/100\n      timeout: 2000"
}

func globalGo() string {
	return "package global\n\nimport \"github.com/panco95/go-garden/core\"\n\nvar (\n\tService *core.Garden\n)\n"
}

func gatewayRpcBase() string {
	return "package rpc\n\nimport \"github.com/panco95/go-garden/core\"\n\ntype Rpc struct {\n\tcore.Rpc\n}\n"
}

func serviceMain(serviceName string) string {
	return strings.Replace("package main\n\nimport (\n\t\"github.com/panco95/go-garden/core\"\n\t\"<>/api\"\n\t\"<>/global\"\n\t\"<>/rpc\"\n)\n\nfunc main() {\n\tglobal.Service = core.New()\n\tglobal.Service.Run(api.Routes, new(rpc.Rpc), nil)\n}\n", "<>", serviceName, 999)
}

func rpcDefineTest() string {
	return "package define\n\ntype TestRpcArgs struct {\n\tPing string\n}\n\ntype TestRpcReply struct {\n\tPong string\n}\n"
}

func rpcBase() string {
	return "package rpc\n\nimport \"github.com/panco95/go-garden/core\"\n\ntype Rpc struct {\n\tcore.Rpc\n}\n"
}

func rpcTest(serviceName string) string {
	return strings.Replace("package rpc\n\nimport (\n\t\"context\"\n\t\"github.com/panco95/go-garden/core\"\n\t\"<>/global\"\n\t\"<>/rpc/define\"\n)\n\nfunc (r *Rpc) TestRpc(ctx context.Context, args *define.TestRpcArgs, reply *define.TestRpcReply) error {\n\tspan := global.Service.StartRpcTrace(ctx, args, \"TestRpc\")\n\n\tglobal.Service.Log(core.InfoLevel, \"Test\", \"Receive a rpc message\")\n\treply.Pong = \"pong\"\n\n\tglobal.Service.FinishRpcTrace(span)\n\treturn nil\n}\n", "<>", serviceName, 999)
}

func apiBase(serviceName string) string {
	return strings.Replace("package api\n\nimport (\n\t\"github.com/gin-gonic/gin\"\n\t\"<>/global\"\n)\n\nfunc Routes(r *gin.Engine) {\n\tr.Use(global.Service.CheckCallSafeMiddleware()) // 调用接口安全验证\n\tr.POST(\"test\", Test)\n}\n", "<>", serviceName, 999)
}

func apiTest(serviceName string) string {
	return strings.Replace("package api\n\nimport (\n\t\"github.com/gin-gonic/gin\"\n\t\"github.com/panco95/go-garden/core\"\n\t\"<>/global\"\n\t\"<>/rpc/define\"\n)\n\nfunc Test(c *gin.Context) {\n\tspan, _ := core.GetSpan(c)\n\n\t// rpc call test\n\targs := define.TestRpcArgs{\n\t\tPing: \"ping\",\n\t}\n\treply := define.TestRpcReply{}\n\t_, _, err := global.Service.CallService(span, \"<>\", \"TestRpc\", nil, &args, &reply)\n\tif err != nil {\n\t\tspan.SetTag(\"CallService\", err)\n\t\tcore.Resp(c, core.HttpFail, 0, core.InfoServerError, nil)\n\t\treturn\n\t}\n\n\tcore.Resp(c, core.HttpOk, 0, core.InfoSuccess, core.MapData{\n\t\t\"pong\": reply.Pong,\n\t})\n}\n", "<>", serviceName, 999)
}
