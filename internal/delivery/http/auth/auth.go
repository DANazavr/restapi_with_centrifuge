package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/DANazavr/RATest/internal/common/meta"
	delivery "github.com/DANazavr/RATest/internal/delivery/http"
	"github.com/DANazavr/RATest/internal/domain"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/services"
	"github.com/golang-jwt/jwt/v4"
)

type MiddlewareAuth struct {
	ctx         context.Context
	logger      *log.Log
	authService *services.AuthService
}

func NewMiddlewareAuth(ctx context.Context, logger *log.Log, as *services.AuthService) *MiddlewareAuth {
	return &MiddlewareAuth{
		ctx:         ctx,
		logger:      logger.WithComponent("auth/authMiddleware"),
		authService: as,
	}
}

func (am *MiddlewareAuth) Auth(next http.Handler) http.Handler {
	return am.AuthMiddleware(next)
}

func (am *MiddlewareAuth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			delivery.HendleError(w, r, http.StatusUnauthorized, domain.ErrEmptyToken)
			am.logger.Warnf(r.Context(), "Authorization header is empty")
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			delivery.HendleError(w, r, http.StatusUnauthorized, domain.ErrInvalidAuthHeader)
			am.logger.Warnf(r.Context(), "Invalid Authorization header format: %s", authHeader)
			return
		}

		token, err := am.authService.ParseToken(tokenParts[1])
		if err != nil {
			delivery.HendleError(w, r, http.StatusUnauthorized, err)
			am.logger.Warnf(r.Context(), "Failed to parse token: %v", err)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			tokenType, ok := claims["type"].(string)
			if !ok || tokenType != "access" {
				delivery.HendleError(w, r, http.StatusUnauthorized, domain.ErrInvalidToken)
				am.logger.Warnf(r.Context(), "Invalid token type: %v", tokenType)
				return
			}
			ctxWithUser := context.WithValue(r.Context(), meta.UserIDKey, claims["sub"])
			r = r.WithContext(ctxWithUser)
			am.logger.Infof(ctxWithUser, "Authenticated user ID: %v", claims["sub"])
		}
		next.ServeHTTP(w, r)
	})
}
