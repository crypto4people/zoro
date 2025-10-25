package crypto

import (
	"crypto/rand"
	"crypto/sha3"
	"errors"
	"io"
	"unsafe"

	"golang.org/x/crypto/nacl/secretbox"
)

const CipherSize = 24 + secretbox.Overhead

func Seal(plain []byte, password string) []byte {
	nonce := [24]byte{}
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		panic(err)
	}
	key := sha3.Sum256(unsafe.Slice(unsafe.StringData(password), len(password)))
	return secretbox.Seal(nonce[:], plain, &nonce, &key)
}

func Open(cipher []byte, password string, fn func(plain []byte)) bool {
	if len(cipher) <= CipherSize {
		panic(errors.New("invalid cipher"))
	}
	key := sha3.Sum256(unsafe.Slice(unsafe.StringData(password), len(password)))
	nonce := [24]byte(cipher[:24])
	data, ok := secretbox.Open(nil, cipher[24:], &nonce, &key)
	if ok {
		fn(data)
	}
	return ok
}
