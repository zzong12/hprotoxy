package codec

import (
	"bytes"
	"compress/gzip"
)

type gzipCodec struct {
}

func (c *gzipCodec) Name() string {
	return "gzip"
}

func (c *gzipCodec) Encode(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(data)
	gz.Flush()
	gz.Close()
	return buf.Bytes(), nil
}

func (c *gzipCodec) Decode(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	buf.ReadFrom(gz)
	return buf.Bytes(), nil
}
