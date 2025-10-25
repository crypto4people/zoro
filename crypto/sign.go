package crypto

import (
	"crypto/ecdsa"
	"slices"

	"github.com/ethereum/go-ethereum/crypto"
)

const SigSize = 65

func sign(private *ecdsa.PrivateKey, msg []byte) []byte {
	sig, err := crypto.Sign(crypto.Keccak256(msg), private)
	if err != nil {
		panic(err)
	}
	return slices.Concat(sig, msg)
}

// recover is not safe to be used externally
func recover(box []byte) (pub *ecdsa.PublicKey, msg []byte) {
	if len(box) <= SigSize {
		return nil, nil
	}
	sig := box[:SigSize]
	msg = box[SigSize:]
	digest := crypto.Keccak256(msg)
	var err error
	pub, err = crypto.SigToPub(digest, sig)
	if err != nil || pub == nil {
		return nil, nil
	}
	if !crypto.VerifySignature(crypto.FromECDSAPub(pub), digest, sig[:64]) {
		return nil, nil
	}
	return pub, msg
}
