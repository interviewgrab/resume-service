package auth

import (
	"github.com/golang-jwt/jwt"
	"resume-service/internal/model"
	"time"
)

const jwtSecret = "your_jwt_secret"

func GenerateJWTToken(user model.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID.Hex(),
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	})

	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
