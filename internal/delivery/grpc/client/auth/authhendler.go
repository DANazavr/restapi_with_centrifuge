package auth

import (
	"context"
	"encoding/json"
	"net/http"

	delivery "github.com/DANazavr/RATest/internal/delivery/http"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	ctx    context.Context
	logger *log.Log
	client auth.AuthClient
}

func NewAuthClient(ctx context.Context, logger *log.Log) (*AuthClient, error) {
	conn, err := grpc.NewClient(":8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &AuthClient{
		ctx:    ctx,
		logger: logger.WithComponent("grpc/auth/AuthClient"),
		client: auth.NewAuthClient(conn),
	}, nil
}

func (c *AuthClient) Register() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			delivery.HendleError(w, r, http.StatusBadRequest, err)
			return
		}
		resp, err := c.client.Register(ctx, &auth.RegisterRequest{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
			Role:     req.Role,
		})
		if err != nil {
			delivery.HendleError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		delivery.HendleRespond(w, r, http.StatusCreated, resp)
	}
}

func (c *AuthClient) Login() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			delivery.HendleError(w, r, http.StatusBadRequest, err)
			return
		}
		resp, err := c.client.Login(c.ctx, &auth.LoginRequest{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			delivery.HendleError(w, r, http.StatusUnauthorized, err)
			return
		}
		delivery.HendleRespond(w, r, http.StatusOK, &response{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
		})
	}
}

func (c *AuthClient) TokenRefresh() http.HandlerFunc {
	type request struct {
		RefreshToken string `json:"refresh_token"`
	}

	type response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			delivery.HendleError(w, r, http.StatusBadRequest, err)
			return
		}
		resp, err := c.client.TokenRefresh(c.ctx, &auth.TokenRefreshRequest{
			RefreshToken: req.RefreshToken,
		})
		if err != nil {
			delivery.HendleError(w, r, http.StatusUnauthorized, err)
			return
		}
		delivery.HendleRespond(w, r, http.StatusOK, &response{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
		})
	}
}
