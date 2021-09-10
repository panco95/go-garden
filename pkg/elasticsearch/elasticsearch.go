package elasticsearch

import (
	"context"
	"github.com/olivere/elastic/v7"
)

var esClient *elastic.Client

// EsConnect 初始化连接Elasticsearch
func EsConnect(address string) error {
	var err error
	esClient, err = elastic.NewClient(
		elastic.SetURL(address),
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
	put, err := esClient.Index().
		Index(index).
		BodyString(body).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	return put, nil
}

// GetClient 获取ES客户端
func GetClient() *elastic.Client {
	return esClient
}
