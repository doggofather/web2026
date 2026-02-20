package ds

import (
	"github.com/golang-jwt/jwt"
)

// Если вы используете библиотеку google/uuid

type JWTClaims struct {
	jwt.StandardClaims
	UserUUID uint     `json:"user_uuid"`
	Scopes   []string `json:"scopes"`
}
