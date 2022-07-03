package serializer

import (
	jsoniter "github.com/json-iterator/go"
	"io"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func NewDecoder(r io.Reader) *jsoniter.Decoder {
	return json.NewDecoder(r)
}

func NewEncoder(w io.Writer) *jsoniter.Encoder {
	return json.NewEncoder(w)
}
