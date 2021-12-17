package main

import (
	"fmt"
	"github.com/panco95/go-garden/core"
)

func main() {
	data := map[string]interface{}{
		"menu1": 1000,
		"menu2": 2000,
		"menu3": 3000,
	}
	body, err := core.PushGateway("test1111", data)
	if err != nil {
		panic(err)
	}
	fmt.Println(body)
	//global.Garden = core.New()
	//global.Garden.Run(api.Routes, new(rpc.Rpc), global.Garden.CheckCallSafeMiddleware)
}
