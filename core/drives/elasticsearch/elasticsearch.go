package elasticsearch

import (
	"context"
	"github.com/olivere/elastic/v7"
)

var client *elastic.Client

func Client() *elastic.Client {
	return client
}

func Connect(address string) error {
	var err error
	client, err = elastic.NewClient(
		elastic.SetURL(address),
		elastic.SetSniff(false),
	)
	if err != nil {
		return err
	}
	return nil
}

func Put(index, body string) (*elastic.IndexResponse, error) {
	ctx := context.Background()
	put, err := client.Index().
		Index(index).
		BodyString(body).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	return put, nil
}
