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
	"encoding/json"

	jwt "github.com/golang-jwt/jwt/v5"
)

type jwtService struct{}

func NewJwtService() Signer {
	return &jwtService{}
}

func (s *jwtService) Validate(tokenString string, secret []byte) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return token, nil
}

func (s *jwtService) ValidateTarget(tokenString string, secret []byte, target any) error {
	token, err := s.Validate(tokenString, secret)
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ErrInvalidTokenClaims
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(claimsJSON, target); err != nil {
		return ErrTokenMapping
	}

	return nil
}

func (s *jwtService) Create(claims jwt.Claims, secret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
