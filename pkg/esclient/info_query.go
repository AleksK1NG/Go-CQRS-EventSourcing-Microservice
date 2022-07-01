package esclient

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func Info(ctx context.Context, transport esapi.Transport) (*esapi.Response, error) {
	infoRequest := &esapi.InfoRequest{Pretty: true}
	return infoRequest.Do(ctx, transport)
}
