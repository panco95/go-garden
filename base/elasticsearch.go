package base

import (
	"context"
	"github.com/olivere/elastic/v7"
)

var EsClient *elastic.Client

type es struct{}

var Es es

//InitEs 初始化连接Elasticsearch
func InitEs(address string) error {
	var err error
	EsClient, err = elastic.NewClient(
		elastic.SetURL("http://"+address),
		elastic.SetSniff(false),
	)
	if err != nil {
		return err
	}
	return nil
}

//Put 存储数据到es
func (es) Put(index, body string) (*elastic.IndexResponse, error) {
	ctx := context.Background()
	put, err := EsClient.Index().
		Index(index).
		BodyString(body).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	return put, nil
}
