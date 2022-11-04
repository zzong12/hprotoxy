package codec

import "net/url"

type urlCodec struct {
}

func (c *urlCodec) Name() string {
	return "url"
}

func (c *urlCodec) Encode(data []byte) ([]byte, error) {
	return []byte(url.QueryEscape(string(data))), nil
}

func (c *urlCodec) Decode(data []byte) ([]byte, error) {
	res, err := url.QueryUnescape(string(data))
	if err != nil {
		return nil, err
	}
	return []byte(res), nil
}