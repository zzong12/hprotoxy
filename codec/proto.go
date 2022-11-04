package codec

import (
	"bytes"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/zzong12/hprotoxy/loader"
)

type protoCodec struct {
	Req string `json:"req"`
	Res string `json:"res"`
}

func (c *protoCodec) Name() string {
	return "pb"
}

func (c *protoCodec) Encode(data []byte) ([]byte, error) {
	desc, err := loader.GetLocalLoader().GetMessageDescriptor(c.Req)
	if err != nil {
		return nil, err
	}
	msg := dynamic.NewMessage(desc)
	err = jsonpb.UnmarshalString(string(data), msg)
	if err != nil {
		return nil, err
	}
	return msg.Marshal()
}

func (c *protoCodec) Decode(data []byte) ([]byte, error) {
	desc, err := loader.GetLocalLoader().GetMessageDescriptor(c.Res)
	if err != nil {
		return nil, err
	}
	msg := dynamic.NewMessage(desc)
	if err = msg.Unmarshal(data); err != nil {
		return nil, err
	}
	marshaler := jsonpb.Marshaler{
		EmitDefaults: true,
	}
	buf := bytes.NewBuffer(nil)

	if err = marshaler.Marshal(buf, msg); err != nil {
		return nil, fmt.Errorf("Failed to marshal response: %v", err)
	}
	return buf.Bytes(), nil
}
