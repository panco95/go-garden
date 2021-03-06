package core

import (
	"errors"
	"fmt"

	clientV3 "go.etcd.io/etcd/client/v3"
)

// unSafeList is used keys
var unSafeList = map[string]interface{}{
	"etcd": nil,
}

// setSafe keys
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

//Set instance by name, not support default name like 'etcd'
func (g *Garden) Set(name string, val interface{}) error {
	if _, ok := unSafeList[name]; ok {
		return errors.New("Cant's set unsafe name! ")
	}
	g.container.Store(name, val)
	return nil
}

//GetEtcd instance to performing etcd operations
func (g *Garden) GetEtcd() (*clientV3.Client, error) {
	res, err := g.Get("etcd")
	if err != nil {
		return nil, err
	}
	return res.(*clientV3.Client), nil
}
