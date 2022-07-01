package esclient

import (
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func GetByID[T any, V GetResponse[T]](ctx context.Context, transport esapi.Transport, index, documentID string) (*V, error) {
	request := esapi.GetRequest{
		Index:      index,
		DocumentID: documentID,
		Pretty:     true,
	}

	response, err := request.Do(ctx, transport)
	if err != nil {
		return new(V), err
	}
	defer response.Body.Close()

	var getResponse V
	if err := json.NewDecoder(response.Body).Decode(&getResponse); err != nil {
		return new(V), err
	}

	return &getResponse, nil
}
