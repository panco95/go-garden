package core

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	clientV3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
)

var unSafeList = map[string]interface{}{
	"db":    nil,
	"redis": nil,
	"etcd":  nil,
}

func (g *Garden) Get(name string) (interface{}, error) {
	if res, ok := g.container.Load(name); ok {
		return res, nil
	}
	return nil, errors.New(fmt.Sprintf("Not found %s from container! ", name))
}

func (g *Garden) Set(name string, val interface{}) error {
	if _, ok := unSafeList[name]; ok {
		return errors.New("Cant's set unsafe name! ")
	}
	g.container.Store(name, val)
	return nil
}

func (g *Garden) GetDb() *gorm.DB {
	res, _ := g.Get("db")
	return res.(*gorm.DB)
}

func (g *Garden) GetRedis() *redis.Client {
	res, _ := g.Get("redis")
	return res.(*redis.Client)
}

func (g *Garden) GetEtcd() *clientV3.Client {
	res, _ := g.Get("etcd")
	return res.(*clientV3.Client)
}
