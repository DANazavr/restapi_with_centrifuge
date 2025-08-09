package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/DANazavr/RATest/internal/domain"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/golang-jwt/jwt/v4"
)

type AuthService struct {
	ctx    context.Context
	logger *log.Log
}

func NewAuthService(ctx context.Context, logger *log.Log) *AuthService {
	return &AuthService{
		ctx:    ctx,
		logger: logger.WithComponent("authservice"),
	}
}

func (am *AuthService) GenerateAccessToken(userID int, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  fmt.Sprintf("%d", userID),
		"exp":  jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		"iat":  jwt.NewNumericDate(time.Now()),
		"role": role,
		"type": "access",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("SECRET")))
}

func (am *AuthService) GenerateRefreshToken(userID int, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  fmt.Sprintf("%d", userID),
		"exp":  jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)), // 30 days
		"iat":  jwt.NewNumericDate(time.Now()),
		"role": role,
		"type": "refresh",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("SECRET")))
}

func (am *AuthService) GenerateTokens(userID int, role string) (string, string, error) {
	accessToken, err := am.GenerateAccessToken(userID, role)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := am.GenerateRefreshToken(userID, role)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (am *AuthService) ParseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidSigningMethod
		}
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token, nil
}
