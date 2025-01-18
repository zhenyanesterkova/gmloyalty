package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/zhenyanesterkova/gmloyalty/internal/config"
)

var (
	ErrNoValidToken = errors.New("token is not valid")
)

type SessionsJWT struct {
	secret   []byte
	tokenExp time.Duration
}

type SessionJWTClaims struct {
	jwt.RegisteredClaims
	UserID int `json:"uid"`
}

func NewSessionsJWT(conf config.JWTConfig) *SessionsJWT {
	return &SessionsJWT{
		secret:   []byte(conf.SecretKey),
		tokenExp: conf.TokenExp,
	}
}

func (sm *SessionsJWT) Check(tokenJWT string) (int, error) {
	claims := &SessionJWTClaims{}
	token, err := jwt.ParseWithClaims(tokenJWT, claims,
		func(t *jwt.Token) (interface{}, error) {
			return sm.secret, nil
		})
	if err != nil {
		return 0, fmt.Errorf("failed parse token: %w", err)
	}

	if !token.Valid {
		return 0, ErrNoValidToken
	}

	return claims.UserID, nil
}

func (sm *SessionsJWT) Create(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, SessionJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(sm.tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString(sm.secret)
	if err != nil {
		return "", fmt.Errorf("failed create JWT token: %w", err)
	}

	return tokenString, nil
}
