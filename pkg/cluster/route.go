package cluster

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go-ms/pkg/base/global"
	"io/ioutil"
	"net/http"
)

func GatewayRoute(r *gin.Engine) {
	r.Any("api/:server/:action", func(c *gin.Context) {
		server := c.Param("server")
		action := c.Param("action")
		code, data, err := CallService(server, action)
		if err != nil {
			global.Logger.Debugf("[CallService] %s", err.Error())
			c.JSON(code, global.Any{
				"code":    code,
				"message": "ServerError",
				"data":    nil,
			})
			return
		}
		var r map[string]interface{}
		json.Unmarshal([]byte(data), &r)
		fmt.Println(r)
		c.JSON(code, global.Any{
			"code":    code,
			"message": "success",
			"data":    r["data"],
		})
	})
}

func CallService(serverName, action string) (int, string, error) {
	route := viper.GetString(serverName + "." + action)
	if len(route) == 0 {
		return http.StatusNotFound, "", nil
	}
	serverAddr := chooseServer(serverName)
	res, err := http.Get("http://" + serverAddr + route)
	if err != nil {
		return http.StatusInternalServerError, "", err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return http.StatusOK, string(body), nil
}

func chooseServer(serverName string) string {
	return AnalyzeHttpAddr(serverName, 0)
}
