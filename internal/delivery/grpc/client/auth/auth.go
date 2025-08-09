package auth

import (
	"net/http"

	delivery "github.com/DANazavr/RATest/internal/delivery/http"
	"github.com/DANazavr/RATest/internal/domain"
	"google.golang.org/grpc/metadata"
)

const (
	AuthorizationKey = "authorization"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Извлекаем заголовок Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			delivery.HendleError(w, r, http.StatusUnauthorized, domain.ErrEmptyToken)
			return
		}

		// Добавляем в контекст как метаданные gRPC
		md := metadata.New(map[string]string{AuthorizationKey: authHeader})
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Продолжаем цепочку с обновлённым контекстом
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
