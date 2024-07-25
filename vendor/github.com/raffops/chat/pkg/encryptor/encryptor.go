package encryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

type Encryptor interface {
	Encode(b []byte) string
	Decode(s string) ([]byte, error)
	Encrypt(text, MySecret string) (string, error)
	Decrypt(text, MySecret string) (string, error)
}

var iv = []byte("my16digitIvKey12")

// https://medium.com/insiderengineering/aes-encryption-and-decryption-in-golang-php-and-both-with-full-codes-ceb598a34f41
type defaultEncryptor struct{}

func (c defaultEncryptor) Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func (c defaultEncryptor) Decode(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("error decoding session payload")
	}
	return data, nil
}

func (c defaultEncryptor) Encrypt(text, MySecret string) (string, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return "", err
	}
	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, iv)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return c.Encode(cipherText), nil
}

func (c defaultEncryptor) Decrypt(text, MySecret string) (string, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return "", err
	}
	cipherText, err := c.Decode(text)
	if err != nil {
		return "", err
	}
	cfb := cipher.NewCFBDecrypter(block, iv)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}

func NewDefaultEncryptor() Encryptor {
	return &defaultEncryptor{}
}
