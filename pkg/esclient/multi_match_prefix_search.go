package esclient

import (
	"bytes"
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es/serializer"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type MultiMatch struct {
	Fields []string `json:"fields"`
	Query  string   `json:"query"`
	Type   string   `json:"type"`
}

type MultiMatchQuery struct {
	MultiMatch MultiMatch `json:"multi_match"`
}

type MultiMatchSearchQuery struct {
	Query MultiMatchQuery `json:"query"`
	Sort  []any           `json:"sort"`
}

func SearchMultiMatchPrefix[T any](ctx context.Context, transport esapi.Transport, request SearchMatchPrefixRequest) (*SearchListResponse[T], error) {
	searchQuery := make(map[string]any, 10)
	matchPrefix := make(map[string]any, 10)
	for _, field := range request.Fields {
		matchPrefix[field] = request.Term
	}

	matchSearchQuery := MultiMatchSearchQuery{
		Sort: []interface{}{"_score", request.SortMap},
		Query: MultiMatchQuery{
			MultiMatch: MultiMatch{
				Fields: request.Fields,
				Query:  request.Term,
				Type:   "phrase_prefix",
			}}}

	//searchQuery["query"] = map[string]any{
	//	"multi_match": map[string]any{
	//		"fields": request.Fields,
	//		"query":  request.Term,
	//		"type":   "phrase_prefix",
	//	},
	//}

	if request.SortMap != nil {
		searchQuery["sort"] = []interface{}{"_score", request.SortMap}
		//searchQuery["sort"] = []interface{}{"_score", map[string]interface{}{"age": "asc"}}
	}

	queryBytes, err := serializer.Marshal(&matchSearchQuery)
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
	err = serializer.NewDecoder(response.Body).Decode(&hits)
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
