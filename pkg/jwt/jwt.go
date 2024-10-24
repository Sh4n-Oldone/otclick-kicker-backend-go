package jwt

import (
	// "encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

func NewToken(user entity.User, duration time.Duration, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": user.ID,
		"email": user.Email,
		"role_name": user.Role.Name,
		"team_id": user.Team.ID,
		"token_exp": time.Now().Add(duration).Unix(), //TODO возможно стоит ключ "token_exp" вынести куда-то
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
