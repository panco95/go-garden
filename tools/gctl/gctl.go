package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	name  string
	class string
)

func main() {
	flag.StringVar(&name, "name", "", "name")
	flag.StringVar(&class, "class", "", "gateway or service")
	flag.Parse()
	if len(name) == 0 {
		log.Fatal("-name command is empty, please input --help")
	}
	if len(class) == 0 {
		log.Fatal("-class command is empty, please input --help")
	}
	if class != "gateway" && class != "service" {
		log.Fatal("-class just support gateway or service")
	}

	createDir("./" + name)
	createDir("./" + name + "/configs")

	mainGo := MainGoService
	if class == "gateway" {
		mainGo = MainGoGateway
	}
	createFile("./"+name+"/main.go", mainGo)

	configYml := ConfigYml
	configYml = strings.Replace(configYml, "replace-name", name, 1)
	createFile("./"+name+"/configs/config.yml", configYml)

	createFile("./"+name+"/configs/routes.yml", RoutesYml)

	cmdRun(name, "go", "mod", "init", name)
	cmdRun(name, "go", "mod", "tidy")

	log.Printf("Create service success, dir: ./" + name)
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
		log.Fatal("dir is exists!")
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

func cmdRun(name string, command string, param ...string) {
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

const MainGoGateway = `package main

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
)

var service *core.Garden

func main() {
	service = core.New()
	service.Run(service.GatewayRoute, auth)
}

// Customize the auth middleware
func auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// before logic
		c.Next()
		// after logic
	}
}
`

const MainGoService = `package main

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
)

var service *core.Garden

func main() {
	service = core.New()
	service.Run(route, nil)
}

func route(r *gin.Engine) {
	r.Use(service.CheckCallSafeMiddleware())
	r.POST("test", test)
}

func test(c *gin.Context) {
}
`

const ConfigYml = `service:
  debug: true
  serviceName: replace-name
  listenOut: true
  listenPort: 8080
  callKey: garden
  callRetry: 50/80/100
  etcdKey: garden
  etcdAddress:
    - 127.0.0.1:2379
  zipkinAddress: http://127.0.0.1:9411/api/v2/spans
  amqpAddress: amqp://guest:guest@127.0.0.1:5672

config:
  map:
    a: 1
    b: 2
  int: 1
  intSlice:
    - 1
    - 2
    - 3
  string: hello
  stringSlice:
    - a
    - b
`

const RoutesYml = `routes:
  user:
    login:
      type: out
      path: /login
      limiter: 5/10000
      fusing: 5/1000
      timeout: 2000
    exists:
      type: in
      path: /exists
      limiter: 5/10000
      fusing: 5/1000
      timeout: 2000
  pay:
    order:
      type: out
      path: /order
      limiter: 5/10000
      fusing: 5/1000
      timeout: 2000
`
