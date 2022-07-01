package esclient

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func Delete(ctx context.Context, transport esapi.Transport, index, documentID string) (*esapi.Response, error) {
	request := esapi.DeleteRequest{
		Index:      index,
		DocumentID: documentID,
		Refresh:    "true",
	}

	return request.Do(ctx, transport)
}
