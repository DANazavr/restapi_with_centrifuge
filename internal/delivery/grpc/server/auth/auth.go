package auth

import (
	"context"
	"strings"

	"github.com/DANazavr/RATest/internal/common/meta"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/services"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type InterceptorAuth struct {
	ctx         context.Context
	logger      *log.Log
	authService *services.AuthService
}

func NewInterceptorAuth(ctx context.Context, logger *log.Log, as *services.AuthService) *InterceptorAuth {
	return &InterceptorAuth{
		ctx:         ctx,
		logger:      logger.WithComponent("grpc/auth/InterceptorAuth"),
		authService: as,
	}
}

func (ia *InterceptorAuth) AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod == "/ratest.auth.Auth/Login" ||
		info.FullMethod == "/ratest.auth.Auth/RefreshToken" ||
		info.FullMethod == "/ratest.auth.Auth/Register" ||
		info.FullMethod == "/notification.Notification/Publish" ||
		info.FullMethod == "/notification.Notification/Broadcast" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		ia.logger.Errorf(ctx, "Failed to get metadata")
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	// Получаем токен из заголовка
	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		ia.logger.Errorf(ctx, "Authorization header is missing")
		return nil, status.Error(codes.Unauthenticated, "authorization token is required")
	}
	header := strings.Split(authHeader[0], " ")
	if len(header) != 2 || header[0] != "Bearer" {
		ia.logger.Warnf(ctx, "Invalid Authorization header format: %s", authHeader)
		return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
	}

	token, err := ia.authService.ParseToken(header[1])
	if err != nil {
		ia.logger.Warnf(ctx, "Failed to parse token: %v", err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenType, ok := claims["type"].(string)
		if !ok || tokenType != "access" {
			ia.logger.Warnf(ctx, "Invalid token type: %v", tokenType)
			return nil, status.Error(codes.Unauthenticated, "invalid token type")
		}
		ctx = context.WithValue(ctx, meta.UserIDKey, claims["sub"])
		ia.logger.Infof(ctx, "Authenticated user ID: %v", claims["sub"])
	}
	return handler(ctx, req)
}
