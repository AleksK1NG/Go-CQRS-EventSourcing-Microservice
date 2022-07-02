package esclient

// Doc is update request wrapper for any json serializable data
type Doc struct {
	Doc any `json:"doc"`
}

type GetResponse[T any] struct {
	Index       string `json:"_index"`
	Id          string `json:"_id"`
	Version     int    `json:"_version"`
	SeqNo       int    `json:"_seq_no"`
	PrimaryTerm int    `json:"_primary_term"`
	Found       bool   `json:"found"`
	Source      T      `json:"_source"`
}

type EsHits[T any] struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source T `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type SearchMatchPrefixRequest struct {
	Index   []string
	Term    string
	Size    int
	From    int
	Sort    []string
	Fields  []string
	SortMap map[string]any
}

type SearchListResponse[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
}

type SearchOptions struct {
	Size   int
	From   int
	Sort   []string
	Fields []string
}
