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
	createFile("./"+serviceName+"/configs/config.yml", configsYml(serviceName, "true"))
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
	createFile("./"+serviceName+"/configs/config.yml", configsYml(serviceName, "false"))
	createDir("./" + serviceName + "/global")
	createFile("./"+serviceName+"/global/global.go", globalGo())
	createDir("./" + serviceName + "/rpc")
	createFile("./"+serviceName+"/rpc/base.go", rpcBase())
	createFile("./"+serviceName+"/rpc/test.go", rpcTest(serviceName))
	createDir("./" + serviceName + "/rpc/define")
	createFile("./"+serviceName+"/rpc/define/test.go", rpcDefineTest())
	createDir("./" + serviceName + "/api")
	createFile("./"+serviceName+"/api/routes.go", apiRoutes())
	createFile("./"+serviceName+"/api/define.go", apiDefine())
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
	return strings.Replace("package main\n\nimport (\n\t\"github.com/panco95/go-garden/core\"\n\t\"<>/auth\"\n\t\"<>/global\"\n\t\"<>/rpc\"\n)\n\nfunc main() {\n\tglobal.Garden = core.New()\n\tglobal.Garden.Run(global.Garden.GatewayRoute, new(rpc.Rpc), auth.Auth)\n}\n", "<>", serviceName, 999)
}

func gatewayAuth() string {
	return "package auth\n\nimport \"github.com/gin-gonic/gin\"\n\n// Auth Customize the auth middleware\nfunc Auth() gin.HandlerFunc {\n\treturn func(c *gin.Context) {\n\t\t// before logic\n\t\tc.Next()\n\t\t// after logic\n\t}\n}\n"
}

func configsYml(serviceName, httpOut string) string {
	return strings.Replace("service:\n  debug: true\n  serviceName: <>\n  httpOut: "+httpOut+"\n  httpPort: 8080\n  allowCors: true\n  rpcOut: false\n  rpcPort: 9000\n  callKey: garden\n  callRetry: 20/30/50\n  etcdKey: garden\n  etcdAddress:\n    - 127.0.0.1:2379\n  zipkinAddress: http://127.0.0.1:9411/api/v2/spans\n\nconfig:\n", "<>", serviceName, 999)
}

func serviceRoutesYml(serviceName string) string {
	return strings.Replace("routes:\n  <>:\n    test:\n      type: http\n      path: /test\n      limiter: 5/100\n      fusing: 5/100\n      timeout: 2000\n    testrpc:\n      type: rpc\n      limiter: 5/100\n      fusing: 5/100\n      timeout: 2000", "<>", serviceName, 999)
}

func gatewayRoutesYml() string {
	return "routes:\n serviceName:\n    test:\n      type: http\n      path: /test\n      limiter: 5/100\n      fusing: 5/100\n      timeout: 2000\n    testrpc:\n      type: rpc\n      limiter: 5/100\n      fusing: 5/100\n      timeout: 2000"
}

func globalGo() string {
	return "package global\n\nimport \"github.com/panco95/go-garden/core\"\n\nvar (\n\tGarden *core.Garden\n)\n"
}

func gatewayRpcBase() string {
	return "package rpc\n\nimport \"github.com/panco95/go-garden/core\"\n\ntype Rpc struct {\n\tcore.Rpc\n}\n"
}

func serviceMain(serviceName string) string {
	return strings.Replace("package main\n\nimport (\n\t\"github.com/panco95/go-garden/core\"\n\t\"<>/api\"\n\t\"<>/global\"\n\t\"<>/rpc\"\n)\n\nfunc main() {\n\tglobal.Garden = core.New()\n\tglobal.Garden.Run(api.Routes, new(rpc.Rpc), global.Garden.CheckCallSafeMiddleware)\n}\n", "<>", serviceName, 999)
}

func rpcDefineTest() string {
	return "package define\n\ntype TestrpcArgs struct {\n\tPing string\n}\n\ntype TestrpcReply struct {\n\tPong string\n}\n"
}

func rpcBase() string {
	return "package rpc\n\nimport \"github.com/panco95/go-garden/core\"\n\ntype Rpc struct {\n\tcore.Rpc\n}\n"
}

func rpcTest(serviceName string) string {
	return strings.Replace("package rpc\n\nimport (\n\t\"context\"\n\t\"github.com/panco95/go-garden/core\"\n\t\"<>/global\"\n\t\"<>/rpc/define\"\n)\n\nfunc (r *Rpc) Testrpc(ctx context.Context, args *define.TestrpcArgs, reply *define.TestrpcReply) error {\n\tspan := global.Garden.StartRpcTrace(ctx, args, \"testrpc\")\n\n\tglobal.Garden.Log(core.InfoLevel, \"Test\", \"Receive a rpc message\")\n\treply.Pong = \"pong\"\n\n\tglobal.Garden.FinishRpcTrace(span)\n\treturn nil\n}\n", "<>", serviceName, 999)
}

func apiRoutes() string {
	return "package api\n\nimport (\n\t\"github.com/gin-gonic/gin\"\n)\n\nfunc Routes(r *gin.Engine) {\n\tr.POST(\"test\", Test)\n}\n"
}

func apiTest(serviceName string) string {
	return strings.Replace("package api\n\nimport (\n\t\"github.com/gin-gonic/gin\"\n\t\"github.com/panco95/go-garden/core\"\n\t\"<>/global\"\n\t\"<>/rpc/define\"\n)\n\nfunc Test(c *gin.Context) {\n\tspan, _ := core.GetSpan(c)\n\n\t// rpc call test\n\targs := define.TestrpcArgs{\n\t\tPing: \"ping\",\n\t}\n\treply := define.TestrpcReply{}\n\terr := global.Garden.CallRpc(span, \"<>\", \"testrpc\", &args, &reply)\n\tif err != nil {\n\t\tglobal.Garden.Log(core.ErrorLevel, \"rpcCall\", err)\n\t\tspan.SetTag(\"CallService\", err)\n\t\tFail(c, MsgFail)\n\t\treturn\n\t}\n\n\tSuccess(c, MsgOk, core.MapData{\n\t\t\"pong\": reply.Pong,\n\t})\n}\n", "<>", serviceName, 999)
}

func apiDefine() string {
	return "package api\n\nimport (\n\t\"github.com/gin-gonic/gin\"\n\t\"github.com/panco95/go-garden/core\"\n)\n\nconst (\n\tCodeOk   = 1000\n\tCodeFail = 1001\n\n\tMsgOk            = \"Success\"\n\tMsgFail          = \"Server error\"\n\tMsgInvalidParams = \"Invalid params\"\n)\n\nfunc Success(c *gin.Context, msg string, data core.MapData) {\n\tc.JSON(200, core.MapData{\n\t\t\"code\": CodeOk,\n\t\t\"msg\":  msg,\n\t\t\"data\": data,\n\t})\n}\n\nfunc Fail(c *gin.Context, msg string) {\n\tc.JSON(200, core.MapData{\n\t\t\"code\": CodeFail,\n\t\t\"msg\":  msg,\n\t\t\"data\": nil,\n\t})\n}\n"
}
