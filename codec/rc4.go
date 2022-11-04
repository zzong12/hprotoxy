package codec

import "crypto/rc4"

type rc4Codec struct {
	Key string `json:"key"`
	Iv  string `json:"iv"`
}

func (c *rc4Codec) Name() string {
	return "rc4"
}

func (c *rc4Codec) Encode(data []byte) ([]byte, error) {
	return c.encrip(c.Key, data)
}

func (c *rc4Codec) Decode(data []byte) ([]byte, error) {
	return c.encrip(c.Key, data)
}

func (c *rc4Codec) encrip(key string, src []byte) ([]byte, error) {
	cipher, err := rc4.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	encrypt := make([]byte, len(src))
	cipher.XORKeyStream(encrypt, src)
	return encrypt, nil
}
