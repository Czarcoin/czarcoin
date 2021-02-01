// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package encryption

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"

	"czarcoin.org/czarcoin/pkg/czarcoin"
)

// EncryptPath encrypts path with the given key
func EncryptPath(path czarcoin.Path, cipher czarcoin.Cipher, key *czarcoin.Key) (encrypted czarcoin.Path, err error) {
	// do not encrypt empty paths
	if len(path) == 0 {
		return path, nil
	}

	if cipher == czarcoin.Unencrypted {
		return path, nil
	}

	comps := czarcoin.SplitPath(path)
	for i, comp := range comps {
		comps[i], err = encryptPathComponent(comp, cipher, key)
		if err != nil {
			return "", err
		}
		key, err = DeriveKey(key, "path:"+comp)
		if err != nil {
			return "", err
		}
	}
	return czarcoin.JoinPaths(comps...), nil
}

// DecryptPath decrypts path with the given key
func DecryptPath(path czarcoin.Path, cipher czarcoin.Cipher, key *czarcoin.Key) (decrypted czarcoin.Path, err error) {
	if cipher == czarcoin.Unencrypted {
		return path, nil
	}

	comps := czarcoin.SplitPath(path)
	for i, comp := range comps {
		comps[i], err = decryptPathComponent(comp, cipher, key)
		if err != nil {
			return "", err
		}
		key, err = DeriveKey(key, "path:"+comps[i])
		if err != nil {
			return "", err
		}
	}
	return czarcoin.JoinPaths(comps...), nil
}

// DerivePathKey derives the key for the given depth from the given root key.
// This method must be called on an unencrypted path.
func DerivePathKey(path czarcoin.Path, key *czarcoin.Key, depth int) (derivedKey *czarcoin.Key, err error) {
	if depth < 0 {
		return nil, Error.New("negative depth")
	}

	// do not derive key from empty path
	if len(path) == 0 {
		return key, nil
	}

	comps := czarcoin.SplitPath(path)
	if depth > len(comps) {
		return nil, Error.New("depth greater than path length")
	}

	derivedKey = key
	for i := 0; i < depth; i++ {
		derivedKey, err = DeriveKey(derivedKey, "path:"+comps[i])
		if err != nil {
			return nil, err
		}
	}
	return derivedKey, nil
}

// DeriveContentKey derives the key for the encrypted object data using the root key.
// This method must be called on an unencrypted path.
func DeriveContentKey(path czarcoin.Path, key *czarcoin.Key) (derivedKey *czarcoin.Key, err error) {
	comps := czarcoin.SplitPath(path)
	if len(comps) == 0 {
		return nil, Error.New("path is empty")
	}
	derivedKey, err = DerivePathKey(path, key, len(comps))
	if err != nil {
		return nil, err
	}
	derivedKey, err = DeriveKey(derivedKey, "content")
	if err != nil {
		return nil, err
	}
	return derivedKey, nil
}

func encryptPathComponent(comp string, cipher czarcoin.Cipher, key *czarcoin.Key) (string, error) {
	// derive the key for the current path component
	derivedKey, err := DeriveKey(key, "path:"+comp)
	if err != nil {
		return "", err
	}

	// use the derived key to derive the nonce
	mac := hmac.New(sha512.New, derivedKey[:])
	_, err = mac.Write([]byte("nonce"))
	if err != nil {
		return "", Error.Wrap(err)
	}

	nonce := new(czarcoin.Nonce)
	copy(nonce[:], mac.Sum(nil))

	// encrypt the path components with the parent's key and the derived nonce
	cipherText, err := Encrypt([]byte(comp), cipher, key, nonce)
	if err != nil {
		return "", Error.Wrap(err)
	}

	nonceSize := czarcoin.NonceSize
	if cipher == czarcoin.AESGCM {
		nonceSize = AESGCMNonceSize
	}

	// keep the nonce together with the cipher text
	return base64.RawURLEncoding.EncodeToString(append(nonce[:nonceSize], cipherText...)), nil
}

func decryptPathComponent(comp string, cipher czarcoin.Cipher, key *czarcoin.Key) (string, error) {
	if comp == "" {
		return "", nil
	}

	data, err := base64.RawURLEncoding.DecodeString(comp)
	if err != nil {
		return "", Error.Wrap(err)
	}

	nonceSize := czarcoin.NonceSize
	if cipher == czarcoin.AESGCM {
		nonceSize = AESGCMNonceSize
	}

	// extract the nonce from the cipher text
	nonce := new(czarcoin.Nonce)
	copy(nonce[:], data[:nonceSize])

	decrypted, err := Decrypt(data[nonceSize:], cipher, key, nonce)
	if err != nil {
		return "", Error.Wrap(err)
	}

	return string(decrypted), nil
}
