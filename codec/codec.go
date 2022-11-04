package codec

import (
	"encoding/json"
	"errors"
	"strings"
)

type Codec interface {
	Name() string
	Encode(data []byte) ([]byte, error)
	Decode(data []byte) ([]byte, error)
}

func GenCodec(name string, data string) (Codec, error) {
	var cc Codec
	switch name {
	case "pb":
		cc = new(protoCodec)
	case "rc4":
		cc = new(rc4Codec)
	case "url":
		cc = new(urlCodec)
	case "base64":
		cc = new(base64Codec)
	case "aes":
		cc = new(aesCodec)
	case "gzip":
		cc = new(gzipCodec)
	default:
		return nil, errors.New("not found codec")
	}
	return cc, json.Unmarshal([]byte(data), cc)
}

type Codecs []Codec

func ParserCodes(desc string) (Codecs, error) {
	if len(desc) == 0 {
		return nil, errors.New("empty codec desc")
	}
	var cs Codecs
	for _, span := range strings.Split(desc, ";") {
		if len(span) == 0 {
			continue
		}
		idx := strings.Index(span, ":")
		if idx == -1 {
		}
		name := span[:idx]
		data := span[idx+1:]
		c, err := GenCodec(name, data)
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	return cs, nil
}

func (cs Codecs) Inverted() Codecs {
	var res Codecs
	for i := len(cs) - 1; i >= 0; i-- {
		res = append(res, cs[i])
	}
	return res
}

func (cs Codecs) EncodeAll(data []byte) ([]byte, error) {
	for _, c := range cs {
		var err error
		data, err = c.Encode(data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (cs Codecs) DecodeAll(data []byte) ([]byte, error) {
	for _, c := range cs {
		var err error
		data, err = c.Decode(data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}
