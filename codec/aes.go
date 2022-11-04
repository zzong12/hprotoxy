package codec

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

type aesCodec struct {
	Key string `json:"key"`
	Iv  string `json:"iv"`
}

func (c *aesCodec) Name() string {
	return "aes"
}

func (c *aesCodec) Encode(data []byte) ([]byte, error) {
	if 0 == len(data) {
		return []byte{}, errors.New("AES/CBC/PKCS5PADDING encrypt failed, src empty")
	}

	block, err := aes.NewCipher([]byte(c.Key))
	if err != nil {
		return []byte{}, err
	}

	ecbEncoder := cipher.NewCBCEncrypter(block, []byte(c.Iv))
	content := c.PKCS5_padding(data, block.BlockSize())
	if len(content)%aes.BlockSize != 0 {
		return []byte{}, errors.New("AES/CBC/PKCS5PADDING encrypt content not a multiple of the block size")
	}

	encrypted := make([]byte, len(content))
	ecbEncoder.CryptBlocks(encrypted, content)
	return encrypted, nil
}

func (c *aesCodec) Decode(data []byte) ([]byte, error) {
	if 0 == len(data) {
		return []byte{}, errors.New("AES/CBC/PKCS5PADDING decrypt failed, src empty")
	}
	block, err := aes.NewCipher([]byte(c.Key))
	if err != nil {
		return []byte{}, err
	}
	ecbDecoder := cipher.NewCBCDecrypter(block, []byte(c.Iv))
	decrypted := make([]byte, len(data))
	ecbDecoder.CryptBlocks(decrypted, data)

	return c.PKCS5_trimming(decrypted), nil
}

func (c *aesCodec) PKCS5_padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func (c *aesCodec) PKCS5_trimming(encryptText []byte) []byte {
	padding := encryptText[len(encryptText)-1]
	return encryptText[:len(encryptText)-int(padding)]
}
