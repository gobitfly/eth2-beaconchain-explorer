package utils

import (
	"bytes"
	"encoding/json"

	"github.com/klauspost/pgzip"
)

func Compress(d any) ([]byte, error) {
	var b bytes.Buffer
	zw := pgzip.NewWriter(&b)
	err := json.NewEncoder(zw).Encode(d)
	if err != nil {
		return nil, err
	}
	zw.Close()
	return b.Bytes(), nil
}

func Decompress(d []byte, dest any) error {
	var b bytes.Buffer
	zr, err := pgzip.NewReader(&b)
	if err != nil {
		return err
	}
	err = json.NewDecoder(zr).Decode(dest)
	if err != nil {
		return err
	}
	return nil
}
