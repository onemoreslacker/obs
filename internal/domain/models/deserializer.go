package models

import (
	"encoding/json"
	"io"
)

type Deserializer struct{}

func NewDeserializer() *Deserializer {
	return &Deserializer{}
}

func (d *Deserializer) Deserialize(r io.Reader, v any) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}
