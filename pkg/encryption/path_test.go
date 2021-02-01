// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package encryption

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"czarcoin.org/czarcoin/pkg/czarcoin"
)

func TestEncryption(t *testing.T) {
	forAllCiphers(func(cipher czarcoin.Cipher) {
		for i, path := range []czarcoin.Path{
			"",
			"/",
			"//",
			"file.txt",
			"file.txt/",
			"fold1/file.txt",
			"fold1/fold2/file.txt",
			"/fold1/fold2/fold3/file.txt",
		} {
			errTag := fmt.Sprintf("%d. %+v", i, path)

			key := new(czarcoin.Key)
			copy(key[:], randData(czarcoin.KeySize))

			encrypted, err := EncryptPath(path, cipher, key)
			if !assert.NoError(t, err, errTag) {
				continue
			}

			decrypted, err := DecryptPath(encrypted, cipher, key)
			if !assert.NoError(t, err, errTag) {
				continue
			}

			assert.Equal(t, path, decrypted, errTag)
		}
	})
}

func TestDeriveKey(t *testing.T) {
	forAllCiphers(func(cipher czarcoin.Cipher) {
		for i, tt := range []struct {
			path      czarcoin.Path
			depth     int
			errString string
		}{
			{"fold1/fold2/fold3/file.txt", -1, "encryption error: negative depth"},
			{"fold1/fold2/fold3/file.txt", 0, ""},
			{"fold1/fold2/fold3/file.txt", 1, ""},
			{"fold1/fold2/fold3/file.txt", 2, ""},
			{"fold1/fold2/fold3/file.txt", 3, ""},
			{"fold1/fold2/fold3/file.txt", 4, ""},
			{"fold1/fold2/fold3/file.txt", 5, "encryption error: depth greater than path length"},
		} {
			errTag := fmt.Sprintf("%d. %+v", i, tt)

			key := new(czarcoin.Key)
			copy(key[:], randData(czarcoin.KeySize))

			encrypted, err := EncryptPath(tt.path, cipher, key)
			if !assert.NoError(t, err, errTag) {
				continue
			}

			derivedKey, err := DerivePathKey(tt.path, key, tt.depth)
			if tt.errString != "" {
				assert.EqualError(t, err, tt.errString, errTag)
				continue
			}
			if !assert.NoError(t, err, errTag) {
				continue
			}

			shared := czarcoin.JoinPaths(czarcoin.SplitPath(encrypted)[tt.depth:]...)
			decrypted, err := DecryptPath(shared, cipher, derivedKey)
			if !assert.NoError(t, err, errTag) {
				continue
			}

			expected := czarcoin.JoinPaths(czarcoin.SplitPath(tt.path)[tt.depth:]...)
			assert.Equal(t, expected, decrypted, errTag)
		}
	})
}

func forAllCiphers(test func(cipher czarcoin.Cipher)) {
	for _, cipher := range []czarcoin.Cipher{
		czarcoin.Unencrypted,
		czarcoin.AESGCM,
		czarcoin.SecretBox,
	} {
		test(cipher)
	}
}
