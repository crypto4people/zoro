package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
)

func Box(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, message []byte) []byte {
	key := deriveKey(privateKey, publicKey)
	if key == nil {
		return nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil
	}

	nonce := make([]byte, aead.NonceSize()) // 12
	if _, err := rand.Read(nonce); err != nil {
		return nil
	}

	ct := aead.Seal(nil, nonce, message, nil)
	out := make([]byte, 0, len(nonce)+len(ct))
	out = append(out, nonce...)
	out = append(out, ct...)
	return out
}

func Unbox(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, ciphertext []byte) []byte {
	key := deriveKey(privateKey, publicKey)
	if key == nil {
		return nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil
	}

	ns := aead.NonceSize()
	if len(ciphertext) < ns {
		return nil
	}

	nonce := ciphertext[:ns]
	body := ciphertext[ns:]

	plain, err := aead.Open(nil, nonce, body, nil)
	if err != nil {
		return nil
	}
	return plain
}

func hkdfExtract(salt, ikm []byte) []byte {
	if salt == nil {
		salt = make([]byte, sha256.Size)
	}

	mac := hmac.New(sha256.New, salt)
	mac.Write(ikm)
	return mac.Sum(nil)
}

func hkdfExpand(prk, info []byte, length int) []byte {
	var (
		out  = make([]byte, 0, length)
		prev []byte
	)

	for ctr := byte(1); len(out) < length; ctr++ {
		mac := hmac.New(sha256.New, prk)
		mac.Write(prev)
		mac.Write(info)
		mac.Write([]byte{ctr})
		prev = mac.Sum(nil)
		out = append(out, prev...)
	}
	return out[:length]
}

func deriveKey(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) []byte {
	if priv == nil || pub == nil || priv.Curve != pub.Curve {
		return nil
	}
	curve := priv.Curve

	xb, _ := curve.ScalarMult(pub.X, pub.Y, priv.D.Bytes())
	if xb == nil {
		return nil
	}

	size := (curve.Params().BitSize + 7) / 8
	ikm := xb.FillBytes(make([]byte, size))

	prk := hkdfExtract(nil, ikm)
	return hkdfExpand(prk, []byte("ECDSA-Box"), 32)

}
