package db2

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

type gzipCompressor struct {
}

func (gzipCompressor) compress(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(src); err != nil {
		return nil, fmt.Errorf("gzip cannot compress data: %w", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("gzip cannot close writer: %w", err)
	}
	return buf.Bytes(), nil
}

func (gzipCompressor) decompress(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, nil
	}
	zr, err := gzip.NewReader(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("gzip cannot create reader: %w", err)
	}
	data, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("gzip cannot read: %w", err)
	}
	return data, nil
}

type noOpCompressor struct{}

func (n noOpCompressor) compress(src []byte) ([]byte, error) {
	return src, nil
}

func (n noOpCompressor) decompress(src []byte) ([]byte, error) {
	return src, nil
}
