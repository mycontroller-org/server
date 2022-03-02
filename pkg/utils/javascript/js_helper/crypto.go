package javascript_helper

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
)

const (
	KeyCryptoMD5    = "md5"
	KeyCryptoSHA1   = "sha1"
	KeyCryptoSHA256 = "sha256"
	KeyCryptoSHA512 = "sha512"
)

type Crypto struct {
}

func (c *Crypto) MD5HexString(data string) string {
	return fmt.Sprintf("%X", c.MD5Bytes(data))
}

func (c *Crypto) MD5Bytes(data string) [16]byte {
	return md5.Sum([]byte(data))
}

func (c *Crypto) Sha1HexString(data string) string {
	return fmt.Sprintf("%X", c.Sha1Bytes(data))
}

func (c *Crypto) Sha1Bytes(data string) [20]byte {
	return sha1.Sum([]byte(data))
}

func (c *Crypto) Sha256HexString(data string) string {
	return fmt.Sprintf("%X", c.Sha256Bytes(data))
}

func (c *Crypto) Sha256Bytes(data string) [32]byte {
	return sha256.Sum256([]byte(data))
}

func (c *Crypto) Sha512HexString(data string) string {
	return fmt.Sprintf("%X", c.Sha256Bytes(data))
}

func (c *Crypto) Sha512Bytes(data string) [64]byte {
	return sha512.Sum512([]byte(data))
}
