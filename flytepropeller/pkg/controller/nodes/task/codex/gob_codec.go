package codex

import (
	"encoding/gob"
	"io"
)

type GobStateCodec struct {
}

func (GobStateCodec) Encode(v interface{}, b io.Writer) error {
	enc := gob.NewEncoder(b)
	return enc.Encode(v)
}

func (GobStateCodec) Decode(b io.Reader, v interface{}) error {
	dec := gob.NewDecoder(b)
	return dec.Decode(v)
}
