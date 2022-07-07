package esclient

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es/serializer"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func SearchMatchPhrasePrefix[T any](ctx context.Context, transport esapi.Transport, request SearchMatchPrefixRequest) (*SearchListResponse[T], error) {
	searchQuery := make(map[string]any, 10)
	matchPrefix := make(map[string]any, 10)
	for _, field := range request.Fields {
		matchPrefix[field] = request.Term
	}

	searchQuery["query"] = map[string]any{
		"bool": map[string]any{
			"must": map[string]any{
				"match_phrase_prefix": matchPrefix,
			}},
	}

	if request.SortMap != nil {
		searchQuery["sort"] = []interface{}{"_score", request.SortMap}
	}

	queryBytes, err := serializer.Marshal(searchQuery)
	if err != nil {
		return nil, err
	}

	searchRequest := esapi.SearchRequest{
		Index:  request.Index,
		Body:   bytes.NewReader(queryBytes),
		Size:   IntPointer(request.Size),
		From:   IntPointer(request.From),
		Sort:   request.Sort,
		Pretty: true,
	}

	response, err := searchRequest.Do(ctx, transport)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	hits := EsHits[T]{}
	err = json.NewDecoder(response.Body).Decode(&hits)
	if err != nil {
		return nil, err
	}

	responseList := make([]T, len(hits.Hits.Hits))
	for i, source := range hits.Hits.Hits {
		responseList[i] = source.Source
	}

	return &SearchListResponse[T]{
		List:  responseList,
		Total: hits.Hits.Total.Value,
	}, nil
}
