package javascript_helper

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
)

const (
	KeyCryptoMD5    = "md5"
	KeyCryptoSHA1   = "sha1"
	KeyCryptoSHA256 = "sha256"
	KeyCryptoSHA512 = "sha512"
)

func getCrypto() map[string]interface{} {
	return map[string]interface{}{
		KeyCryptoMD5:    md5.Sum,
		KeyCryptoSHA1:   sha1.Sum,
		KeyCryptoSHA256: sha256.Sum256,
		KeyCryptoSHA512: sha512.Sum512,
	}
}
