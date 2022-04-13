package core

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	clientV3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var unSafeList = map[string]interface{}{
	"log":   nil,
	"db":    nil,
	"redis": nil,
	"etcd":  nil,
}

func (g *Garden) setSafe(name string, val interface{}) {
	g.container.Store(name, val)
}

//Get instance by name
func (g *Garden) Get(name string) (interface{}, error) {
	if res, ok := g.container.Load(name); ok {
		return res, nil
	}
	return nil, errors.New(fmt.Sprintf("Not found %s from container! ", name))
}

//Set instance by name, not support default name like 'log','db','redis','etcd'
func (g *Garden) Set(name string, val interface{}) error {
	if _, ok := unSafeList[name]; ok {
		return errors.New("Cant's set unsafe name! ")
	}
	g.container.Store(name, val)
	return nil
}

//GetLog instance to write custom Logs
func (g *Garden) GetLog() (*zap.SugaredLogger, error) {
	res, err := g.Get("log")
	if err != nil {
		return nil, err
	}
	return res.(*zap.SugaredLogger), nil
}

//GetDb instance to performing database operations
func (g *Garden) GetDb() *gorm.DB {
	res, _ := g.Get("db")
	return res.(*gorm.DB)
}

//GetRedis instance to performing redis operations
func (g *Garden) GetRedis() *redis.Client {
	res, _ := g.Get("redis")
	return res.(*redis.Client)
}

//GetEtcd instance to performing etcd operations
func (g *Garden) GetEtcd() *clientV3.Client {
	res, _ := g.Get("etcd")
	return res.(*clientV3.Client)
}
