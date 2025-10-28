package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"io"

	"github.com/crypto4people/zoro/misc"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const SigSize = 65

func sign(private *ecdsa.PrivateKey, digest []byte) []byte {
	sig, err := crypto.Sign(digest, private)
	if err != nil {
		panic(err)
	}
	return sig
}

// recover is not safe to be used externally
func recover(sig []byte, digest []byte) (pub *ecdsa.PublicKey) {
	if len(sig) < 65 {
		return nil
	}
	pub, err := crypto.SigToPub(digest, sig)
	if err == nil && pub != nil && crypto.VerifySignature(crypto.FromECDSAPub(pub), digest[:], sig[:64]) {
		return pub
	}
	return nil
}

func NewChallenge() []byte {
	data := make([]byte, 32)
	misc.Must(io.ReadFull(rand.Reader, data))
	return data
}

const magic = "abgr+mahdi"

func sumChallenge(challenge []byte, network, host string) []byte {
	builder := bytes.Buffer{}
	builder.Grow(len(magic) + len(challenge) + len(network) + len(host))
	builder.WriteString(magic)
	builder.Write(challenge)
	builder.WriteString(network)
	builder.WriteString(host)
	return crypto.Keccak256(builder.Bytes())
}

func SignChallenge(private *ecdsa.PrivateKey, rand []byte, network, host string) []byte {
	return sign(private, sumChallenge(rand, network, host))
}

func VerifyChallenge(challenge []byte, addr common.Address, network, host string, sig []byte) bool {
	pub := recover(sig, sumChallenge(challenge, network, host))
	if pub != nil && addr == crypto.PubkeyToAddress(*pub) {
		return true
	}
	return false
}
