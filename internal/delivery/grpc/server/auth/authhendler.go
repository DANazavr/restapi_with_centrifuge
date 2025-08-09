package auth

import (
	"context"
	"strconv"

	"github.com/DANazavr/RATest/internal/domain/models"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/services"
	"github.com/DANazavr/RATest/protos/gen/go/auth"
	"github.com/DANazavr/RATest/protos/gen/go/notification"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	ctx                context.Context
	logger             *log.Log
	notificationClient notification.NotificationClient
	userService        *services.UserService
	authService        *services.AuthService
	auth.UnimplementedAuthServer
}

func NewAuthServer(ctx context.Context, logger *log.Log, us *services.UserService, as *services.AuthService) *AuthServer {
	conn, err := grpc.NewClient(":8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatalf(ctx, "Failed to connect to notification service: %v", err)
	}
	// defer conn.Close()
	notificationClient := notification.NewNotificationClient(conn)

	return &AuthServer{
		ctx:                ctx,
		logger:             logger.WithComponent("grpc/auth/AuthServer"),
		userService:        us,
		authService:        as,
		notificationClient: notificationClient,
	}
}

func Register(gRPC *grpc.Server, authServer *AuthServer) {
	auth.RegisterAuthServer(gRPC, authServer)
}

func (a *AuthServer) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	u := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	}
	if err := a.userService.UsersCreate(ctx, u); err != nil {
		a.logger.Errorf(ctx, "Failed to create user: %v", err)
		return nil, status.Errorf(status.Code(err), "Failed to register user: %v", err)
	}

	// Publish a notification about the new user registration
	if _, err := a.notificationClient.Publish(ctx, &notification.PublishRequest{
		Channel: "notifications:user#" + strconv.Itoa(u.ID),
		Data: &notification.Data{
			Title:   "New User Registration",
			Message: "User " + u.Username + " has Login successfully.",
		},
	}); err != nil {
		a.logger.Errorf(ctx, "Failed to publish notification: %v", err)
		return nil, status.Errorf(status.Code(err), "Failed to send notification: %v", err)
	}
	return &auth.RegisterResponse{Message: "Registration successful"}, nil
}

func (a *AuthServer) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	u, err := a.userService.UsersGetByUsername(req.Username)
	if err != nil || !a.userService.ComparePassword(u, req.Password) {
		a.logger.Errorf(ctx, "Failed to login user: %v", err)
		return nil, status.Errorf(status.Code(err), "Invalid username or password")
	}

	atoken, rtoken, err := a.authService.GenerateTokens(u.ID, u.Role)
	if err != nil {
		a.logger.Errorf(ctx, "Failed to generate tokens: %v", err)
		return nil, status.Errorf(status.Code(err), "Failed to generate tokens: %v", err)
	}
	return &auth.LoginResponse{AccessToken: atoken, RefreshToken: rtoken}, nil
}

func (a *AuthServer) TokenRefresh(ctx context.Context, req *auth.TokenRefreshRequest) (*auth.TokenRefreshResponse, error) {
	var id int64
	token, err := a.authService.ParseToken(req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(status.Code(err), "Invalid refresh token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenType, ok := claims["type"].(string)
		if !ok || tokenType != "refresh" {
			return nil, status.Errorf(status.Code(err), "Invalid token type: %v", tokenType)
		}
		userID, ok := claims["sub"].(string)
		userid, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			a.logger.Info(a.ctx, "failed to parse userID from token")
			return nil, status.Errorf(status.Code(err), "Invalid user ID in token: %v", userID)
		}
		if !ok {
			return nil, status.Errorf(status.Code(err), "Invalid token claims: %v", claims)
		}
		id = userid
	}

	u, err := a.userService.UsersGetById(int(id))
	if err != nil {
		return nil, status.Errorf(status.Code(err), "User not found: %v", err)
	}

	atoken, rtoken, err := a.authService.GenerateTokens(u.ID, u.Role)
	if err != nil {
		return nil, status.Errorf(status.Code(err), "Failed to generate tokens: %v", err)
	}
	return &auth.TokenRefreshResponse{AccessToken: atoken, RefreshToken: rtoken}, nil
}
