package core

import (
	"errors"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"strconv"
	"strings"
	"time"
)

func retryAnalyze(retry string) ([]int, error) {
	retrySlice := make([]int, 0)
	arr := strings.Split(retry, "/")
	if len(arr) == 0 {
		return []int{}, errors.New("config retry format error")
	}
	for _, sec := range arr {
		s, err := strconv.Atoi(sec)
		if err != nil {
			return []int{}, errors.New("config retry format error")
		}
		retrySlice = append(retrySlice, s)
	}

	retrySlice = append(retrySlice, 0)

	return retrySlice, nil
}

func (g *Garden) retryGo(service, action string, retry []int, nodeIndex int, span opentracing.Span, route routeCfg, request *Request, rpcArgs, rpcReply interface{}) (int, string, error) {
	code := httpOk
	result := infoSuccess
	addr := ""
	var err error

	for i, r := range retry {
		sm := serviceOperate{
			operate:     "incWaiting",
			serviceName: service,
			nodeIndex:   nodeIndex,
		}
		g.serviceManager <- sm

		if route.Type == "http" {
			addr, err = g.getServiceHttpAddr(service, nodeIndex)
			if err != nil {
				code = httpFail
				break
			}
			addr = "http://" + addr + route.Path
			code, result, err = g.requestServiceHttp(span, addr, request, route.Timeout)
		} else if route.Type == "rpc" {
			addr, err = g.getServiceRpcAddr(service, nodeIndex)
			if err != nil {
				code = httpFail
				break
			}
			action = capitalize(action)
			err = rpcCall(span, addr, service, action, rpcArgs, rpcReply, route.Timeout)
			if err != nil {
				code = httpFail
			}
		}

		sm.operate = "decWaiting"
		g.serviceManager <- sm

		if err != nil {
			g.Log(ErrorLevel, "callService", err)
			g.addFusingQuantity(service + "/" + action)

			// call timeout don't retry
			if strings.Contains(err.Error(), "Timeout") || strings.Contains(err.Error(), "deadline") {
				err = errors.New(fmt.Sprintf("Call %s %s %s timeout", route.Type, service, action))
				return code, infoTimeout, err
			}

			// call 404 don't retry
			if code == httpNotFound {
				return code, infoNotFound, err
			}

			if i == len(retry)-1 {
				return code, infoServerError, err
			}
			time.Sleep(time.Millisecond * time.Duration(r))
			continue
		}

		break
	}

	return code, result, err
}
