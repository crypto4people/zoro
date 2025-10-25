package crypto

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/secretbox"
)

const CipherSize = 24 + secretbox.Overhead

func Seal(plain []byte, key *[32]byte) []byte {
	nonce := [24]byte{}
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		panic(err)
	}
	return secretbox.Seal(nonce[:], plain, &nonce, key)
}

func open(cipher []byte, key *[32]byte) []byte {
	if len(cipher) <= CipherSize {
		panic(errors.New("invalid cipher"))
	}
	nonce := [24]byte(cipher[:24])
	data, ok := secretbox.Open(nil, cipher[24:], &nonce, key)
	if ok {
		return data
	}
	return nil
}

func Open(cipher []byte, key *[32]byte, fn func(plain []byte) error) error {
	data := open(cipher, key)
	if data == nil {
		return errors.New("invalid cipher")
	}
	return fn(data)
}
