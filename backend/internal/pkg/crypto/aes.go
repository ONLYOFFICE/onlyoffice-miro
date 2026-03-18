/**
 *
 * (c) Copyright Ascensio System SIA 2025
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"

	pbkdf2 "golang.org/x/crypto/pbkdf2"
)

const (
	pbkdf2Iterations = 4096
	saltSize         = 16
	keySize          = 32
)

type aesCipher struct {
	key []byte
}

func NewAESCipher(key []byte) Cipher {
	return &aesCipher{key: key}
}

func generateRandomSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	return salt, nil
}

func (p *aesCipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", ErrCipherTextEmpty
	}

	salt, err := generateRandomSalt(saltSize)
	if err != nil {
		return "", err
	}

	key := pbkdf2.Key(p.key, salt, pbkdf2Iterations, keySize, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	encrypted := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	ciphertextBytes := append(salt, encrypted...)
	return base64.StdEncoding.EncodeToString(ciphertextBytes), nil
}

func (p *aesCipher) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", ErrCipherTextEmpty
	}

	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	if len(ciphertextBytes) < saltSize {
		return "", ErrCipherTextTooShort
	}

	salt := ciphertextBytes[:saltSize]
	encrypted := ciphertextBytes[saltSize:]
	key := pbkdf2.Key(p.key, salt, pbkdf2Iterations, keySize, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return "", ErrCipherTextTooShort
	}

	nonce, ct := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
