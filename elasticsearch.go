package goms

import (
	"context"
	"github.com/olivere/elastic/v7"
)

var EsClient *elastic.Client

// EsConnect 初始化连接Elasticsearch
func EsConnect(address string) error {
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

//EsPut 存储数据到es
func EsPut(index, body string) (*elastic.IndexResponse, error) {
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
