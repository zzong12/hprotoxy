package codec

import "encoding/base64"

type base64Codec struct {
}

func (c *base64Codec) Name() string {
	return "base64"
}

func (c *base64Codec) Encode(data []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)
	return dst, nil
}

func (c *base64Codec) Decode(data []byte) ([]byte, error) {
	dst, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, err
	}
	return dst, nil
}
