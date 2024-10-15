package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

func NewToken(userID int64, roles []int32, duration time.Duration, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid":       userID,
		"roles":     roles,
		"token_exp": time.Now().Add(duration).Unix(), //TODO возможно стоит ключ "token_exp" вынести куда-то, в api-gateway оно будет нужно в middleware Auth
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func NewRefreshToken() string {
	token := uuid.New()
	return token.String()
}
