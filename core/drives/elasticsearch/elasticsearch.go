package elasticsearch

import (
	"context"
	"github.com/olivere/elastic/v7"
)

var client *elastic.Client

// Client get
func Client() *elastic.Client {
	return client
}

// Connect elasticsearch server
func Connect(address string) error {
	var err error
	client, err = elastic.NewClient(
		elastic.SetURL(address),
		elastic.SetSniff(true),
	)
	if err != nil {
		return err
	}
	return nil
}

// Put doc to index
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
