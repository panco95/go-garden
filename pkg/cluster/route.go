package cluster

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go-ms/pkg/base/global"
	"io/ioutil"
	"net/http"
	"strings"
)

func GatewayRoute(r *gin.Engine) {
	r.POST("api/:server/:action", func(c *gin.Context) {
		server := c.Param("server")
		action := c.Param("action")
		jsonBody := global.Any{}
		c.BindJSON(&jsonBody)
		code, data, err := CallServer(server, action, jsonBody)
		if err != nil {
			c.JSON(code, global.Any{
				"code":    code,
				"message": "ServerError",
				"data":    nil,
			})
			return
		}
		var r map[string]interface{}
		json.Unmarshal([]byte(data), &r)
		c.JSON(code, global.Any{
			"code":    code,
			"message": "success",
			"data":    r["data"],
		})
	})
	r.Any("cluster", func(c *gin.Context) {
		c.JSON(http.StatusOK, global.Any{
			"servers": Servers,
		})
	})
}

func CallServer(serverName, action string, jsonBody global.Any) (int, string, error) {
	route := viper.GetString(serverName + "." + action)
	if len(route) == 0 {
		return http.StatusNotFound, "", nil
	}
	serverAddr, err := chooseServer(serverName)
	if err != nil {
		return http.StatusInternalServerError, "", err
	}

	url := "http://" + serverAddr + route
	contentType := "application/json"
	body, err := json.Marshal(jsonBody)
	if err != nil {
		return http.StatusInternalServerError, "", err
	}
	res, err := http.Post(url, contentType, strings.NewReader(string(body)))
	if err != nil {
		return http.StatusInternalServerError, "", err
	}
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	return http.StatusOK, string(body), nil
}

func chooseServer(serverName string) (string, error) {
	if _, ok := Servers[serverName]; !ok {
		return "", errors.New("Server not found")
	}
	serverHttpAddr, err := AnalyzeHttpAddr(serverName, Servers[serverName].PollNext)
	if err != nil {
		return "", err
	}
	go func() {
		serverNum := len(Servers[serverName].Nodes)
		index := Servers[serverName].PollNext
		ServersLock.Lock()
		if index >= serverNum-1 {
			Servers[serverName].PollNext = 0
		} else {
			Servers[serverName].PollNext = index + 1
		}
		Servers[serverName].RequestFinish++
		ServersLock.Unlock()
	}()
	return serverHttpAddr, nil
}
