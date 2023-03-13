package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cristalhq/jwt/v5"
	"github.com/vasiliyantufev/gophermart/internal/model"
)

type JWT struct {
	signer   *jwt.HSAlg
	verifier jwt.Verifier
}

func NewJwt(key []byte) JWT {

	signer, err := jwt.NewSignerHS(jwt.HS256, key)
	if err != nil {
		panic(err)
	}

	verifier, err := jwt.NewVerifierHS(jwt.HS256, key)
	if err != nil {
		panic(err)
	}

	return JWT{signer: signer, verifier: verifier}
}

func GenerateKey() string {
	b, err := jwt.GenerateRandomBits(512)
	if err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(b)
}

func DecodeKey(key string) []byte {
	b, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		panic(err)
	}

	return b
}

func (j JWT) GenerateToken(user *model.UserPayload) (string, error) {
	b, err := json.Marshal(user)
	if err != nil {
		return "", err
	}

	claims := &jwt.RegisteredClaims{
		Audience: []string{string(b)},
	}

	// create a Builder
	builder := jwt.NewBuilder(j.signer)

	// and build a Token
	token, err := builder.Build(claims)
	if err != nil {
		return "", fmt.Errorf("can't build token: %w", err)
	}

	return token.String(), nil
}

func (j JWT) ParseToken(token string) (model.UserPayload, error) {
	newToken, err := jwt.Parse([]byte(token), j.verifier)
	if err != nil {
		return model.UserPayload{}, fmt.Errorf("can't parse token: %w", err)
	}

	var newClaims jwt.RegisteredClaims
	if err := json.Unmarshal(newToken.Claims(), &newClaims); err != nil {
		return model.UserPayload{}, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	var user model.UserPayload
	if err := json.Unmarshal([]byte(newClaims.Audience[0]), &user); err != nil {
		return model.UserPayload{}, fmt.Errorf("failed to unmarshal UserPayload: %w", err)
	}

	return user, nil
}
