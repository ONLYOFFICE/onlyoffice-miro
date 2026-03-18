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
package document

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-miro/backend/internal/pkg/service"
	"github.com/ONLYOFFICE/onlyoffice-miro/backend/pkg/common"
)

type KeyGenerator interface {
	Generate(context.Context, DocumentConfigurer) (string, error)
}

type modificationKeyGenerator struct {
	logger service.Logger
}

func NewModificationKeyGenerator(logger service.Logger) KeyGenerator {
	return &modificationKeyGenerator{
		logger: logger,
	}
}

func (g *modificationKeyGenerator) Generate(ctx context.Context, configurer DocumentConfigurer) (string, error) {
	g.logger.Debug(ctx, "Generating modification key for document", service.Fields{
		"documentId": configurer.ID(),
		"modifiedAt": configurer.ModifiedAt(),
	})

	hasher := md5.New()
	hasher.Write([]byte(common.Concat(configurer.ID(), configurer.ModifiedAt())))
	hash := hasher.Sum(nil)

	key := base64.URLEncoding.EncodeToString(hash)
	g.logger.Debug(ctx, "Generated modification key", service.Fields{
		"key": key,
	})

	return key, nil
}

type SignatureGenerator interface {
	Sign(key []byte, payload []byte) (string, error)
}

type jwtSignatureGenerator struct {
	logger service.Logger
}

func NewJwtSignatureGenerator(logger service.Logger) SignatureGenerator {
	return &jwtSignatureGenerator{
		logger: logger,
	}
}

func (g *jwtSignatureGenerator) Sign(key []byte, payload []byte) (string, error) {
	g.logger.Debug(context.Background(), "Generating JWT signature", service.Fields{
		"payloadSize": len(payload),
	})

	header := `{"alg":"HS256","typ":"JWT"}`
	hencoded := base64.RawURLEncoding.EncodeToString([]byte(header))

	pencoded, err := encodeClaims(payload)
	if err != nil {
		g.logger.Error(context.Background(), "Failed to encode claims", service.Fields{
			"error": err.Error(),
		})
		return "", err
	}

	token := common.Concat(hencoded, ".", pencoded)
	signature := computeHMAC(token, key)

	jwt := common.Concat(token, ".", signature)
	g.logger.Debug(context.Background(), "Generated JWT token", service.Fields{
		"tokenLength": len(jwt),
	})

	return jwt, nil
}

func encodeClaims(payload []byte) (string, error) {
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		claims = make(map[string]any)
	}

	now := time.Now()
	claims["iat"] = now.Unix()
	claims["exp"] = now.Add(5 * time.Minute).Unix()

	npayload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(npayload), nil
}

func computeHMAC(message string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(message))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
