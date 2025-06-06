package models

import "encoding/json"

type Serializer struct{}

func NewSerializer() *Serializer {
	return &Serializer{}
}

func (s *Serializer) Serialize(v any) ([]byte, error) {
	return json.Marshal(v)
}
