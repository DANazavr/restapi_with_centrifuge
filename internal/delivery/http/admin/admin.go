package admin

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

type MiddlewareAdmin struct {
	ctx         context.Context
	logger      *log.Log
	authService *services.AuthService
}

func NewMiddlewareAdmin(ctx context.Context, logger *log.Log, as *services.AuthService) *MiddlewareAdmin {
	return &MiddlewareAdmin{
		ctx:         ctx,
		logger:      logger.WithComponent("auth/adminMiddleware"),
		authService: as,
	}
}

func (ma *MiddlewareAdmin) Admin(next http.Handler) http.Handler {
	return ma.AdminMiddleware(next)
}

func (ma *MiddlewareAdmin) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			delivery.HendleError(w, r, http.StatusUnauthorized, domain.ErrEmptyToken)
			ma.logger.Warnf(r.Context(), "Authorization header is empty")
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			delivery.HendleError(w, r, http.StatusUnauthorized, domain.ErrInvalidAuthHeader)
			ma.logger.Warnf(r.Context(), "Invalid Authorization header format: %s", authHeader)
			return
		}

		token, err := ma.authService.ParseToken(tokenParts[1])
		if err != nil {
			delivery.HendleError(w, r, http.StatusUnauthorized, err)
			ma.logger.Warnf(r.Context(), "Failed to parse token: %v", err)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			tokenType, ok := claims["type"].(string)
			if !ok || tokenType != "access" {
				delivery.HendleError(w, r, http.StatusUnauthorized, domain.ErrInvalidToken)
				ma.logger.Warnf(context.WithValue(r.Context(), meta.UserIDKey, claims["sub"]), "Invalid token type: %v", tokenType)
				return
			}
			tokenRole, ok := claims["role"].(string)
			if !ok || tokenRole != "admin" {
				delivery.HendleError(w, r, http.StatusUnauthorized, domain.ErrInvalidToken)
				ma.logger.Warnf(context.WithValue(r.Context(), meta.UserIDKey, claims["sub"]), "Unauthorized access attempt with role: %v", tokenRole)
				return
			}
			ctx := context.WithValue(r.Context(), meta.UserIDKey, claims["sub"])
			r = r.WithContext(ctx)
			ma.logger.Infof(ctx, "Admin authenticated with user ID: %v", claims["sub"])
		}

		next.ServeHTTP(w, r)
	})
}
