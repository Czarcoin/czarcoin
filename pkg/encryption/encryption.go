// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package encryption

import (
	"crypto/hmac"
	"crypto/sha512"

	"czarcoin.org/czarcoin/pkg/czarcoin"
)

// AESGCMNonceSize is the size of an AES-GCM nonce
const AESGCMNonceSize = 12

// AESGCMNonce represents the nonce used by the AES-GCM protocol
type AESGCMNonce [AESGCMNonceSize]byte

// ToAESGCMNonce returns the nonce as a AES-GCM nonce
func ToAESGCMNonce(nonce *czarcoin.Nonce) *AESGCMNonce {
	aes := new(AESGCMNonce)
	copy((*aes)[:], nonce[:AESGCMNonceSize])
	return aes
}

// Increment increments the nonce with the given amount
func Increment(nonce *czarcoin.Nonce, amount int64) (truncated bool, err error) {
	return incrementBytes(nonce[:], amount)
}

// Encrypt encrypts data with the given cipher, key and nonce
func Encrypt(data []byte, cipher czarcoin.Cipher, key *czarcoin.Key, nonce *czarcoin.Nonce) (cipherData []byte, err error) {
	// Don't encrypt empty slice
	if len(data) == 0 {
		return []byte{}, nil
	}

	switch cipher {
	case czarcoin.Unencrypted:
		return data, nil
	case czarcoin.AESGCM:
		return EncryptAESGCM(data, key, ToAESGCMNonce(nonce))
	case czarcoin.SecretBox:
		return EncryptSecretBox(data, key, nonce)
	default:
		return nil, ErrInvalidConfig.New("encryption type %d is not supported", cipher)
	}
}

// Decrypt decrypts cipherData with the given cipher, key and nonce
func Decrypt(cipherData []byte, cipher czarcoin.Cipher, key *czarcoin.Key, nonce *czarcoin.Nonce) (data []byte, err error) {
	// Don't decrypt empty slice
	if len(cipherData) == 0 {
		return []byte{}, nil
	}

	switch cipher {
	case czarcoin.Unencrypted:
		return cipherData, nil
	case czarcoin.AESGCM:
		return DecryptAESGCM(cipherData, key, ToAESGCMNonce(nonce))
	case czarcoin.SecretBox:
		return DecryptSecretBox(cipherData, key, nonce)
	default:
		return nil, ErrInvalidConfig.New("encryption type %d is not supported", cipher)
	}
}

// NewEncrypter creates a Transformer using the given cipher, key and nonce to encrypt data passing through it
func NewEncrypter(cipher czarcoin.Cipher, key *czarcoin.Key, startingNonce *czarcoin.Nonce, encryptedBlockSize int) (Transformer, error) {
	switch cipher {
	case czarcoin.Unencrypted:
		return &NoopTransformer{}, nil
	case czarcoin.AESGCM:
		return NewAESGCMEncrypter(key, ToAESGCMNonce(startingNonce), encryptedBlockSize)
	case czarcoin.SecretBox:
		return NewSecretboxEncrypter(key, startingNonce, encryptedBlockSize)
	default:
		return nil, ErrInvalidConfig.New("encryption type %d is not supported", cipher)
	}
}

// NewDecrypter creates a Transformer using the given cipher, key and nonce to decrypt data passing through it
func NewDecrypter(cipher czarcoin.Cipher, key *czarcoin.Key, startingNonce *czarcoin.Nonce, encryptedBlockSize int) (Transformer, error) {
	switch cipher {
	case czarcoin.Unencrypted:
		return &NoopTransformer{}, nil
	case czarcoin.AESGCM:
		return NewAESGCMDecrypter(key, ToAESGCMNonce(startingNonce), encryptedBlockSize)
	case czarcoin.SecretBox:
		return NewSecretboxDecrypter(key, startingNonce, encryptedBlockSize)
	default:
		return nil, ErrInvalidConfig.New("encryption type %d is not supported", cipher)
	}
}

// EncryptKey encrypts keyToEncrypt with the given cipher, key and nonce
func EncryptKey(keyToEncrypt *czarcoin.Key, cipher czarcoin.Cipher, key *czarcoin.Key, nonce *czarcoin.Nonce) (czarcoin.EncryptedPrivateKey, error) {
	return Encrypt(keyToEncrypt[:], cipher, key, nonce)
}

// DecryptKey decrypts keyToDecrypt with the given cipher, key and nonce
func DecryptKey(keyToDecrypt czarcoin.EncryptedPrivateKey, cipher czarcoin.Cipher, key *czarcoin.Key, nonce *czarcoin.Nonce) (*czarcoin.Key, error) {
	plainData, err := Decrypt(keyToDecrypt, cipher, key, nonce)
	if err != nil {
		return nil, err
	}

	var decryptedKey czarcoin.Key
	copy(decryptedKey[:], plainData)

	return &decryptedKey, nil
}

// DeriveKey derives new key from the given key and message using HMAC-SHA512
func DeriveKey(key *czarcoin.Key, message string) (*czarcoin.Key, error) {
	mac := hmac.New(sha512.New, key[:])
	_, err := mac.Write([]byte(message))
	if err != nil {
		return nil, Error.Wrap(err)
	}

	derived := new(czarcoin.Key)
	copy(derived[:], mac.Sum(nil))

	return derived, nil
}
