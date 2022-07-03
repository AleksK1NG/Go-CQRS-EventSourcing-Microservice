package esclient

import (
	"bytes"
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es/serializer"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/pkg/errors"
)

func Index(ctx context.Context, transport esapi.Transport, index, documentID string, v any) (*esapi.Response, error) {
	reqBytes, err := serializer.Marshal(v)
	if err != nil {
		return nil, errors.Wrap(err, "json.Marshal")
	}

	request := esapi.IndexRequest{
		Index:      index,
		DocumentID: documentID,
		Body:       bytes.NewBuffer(reqBytes),
	}

	return request.Do(ctx, transport)
}
