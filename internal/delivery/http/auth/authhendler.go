package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	delivery "github.com/DANazavr/RATest/internal/delivery/http"
	"github.com/DANazavr/RATest/internal/domain"
	"github.com/DANazavr/RATest/internal/domain/models"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/services"
	"github.com/DANazavr/RATest/internal/store"
	"github.com/golang-jwt/jwt/v4"
)

type AuthHendler struct {
	ctx         context.Context
	logger      *log.Log
	userService *services.UserService
	authService *services.AuthService
}

func NewAuthHendler(ctx context.Context, logger *log.Log, store store.Store, us *services.UserService, as *services.AuthService) *AuthHendler {
	return &AuthHendler{
		ctx:         ctx,
		logger:      logger.WithComponent("auth/authHendler"),
		userService: us,
		authService: as,
	}
}

func (h *AuthHendler) HandleRegister() http.HandlerFunc {
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
		u := &models.User{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
			Role:     req.Role,
		}
		if err := h.userService.UsersCreate(ctx, u); err != nil {
			delivery.HendleError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		delivery.HendleRespond(w, r, http.StatusCreated, u)
	}
}

func (h *AuthHendler) HandleLogin() http.HandlerFunc {
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
		u, err := h.userService.UsersGetByUsername(req.Username)
		if err != nil || !h.userService.ComparePassword(u, req.Password) {
			delivery.HendleError(w, r, http.StatusUnauthorized, domain.ErrIncorectUsernameOrPassword)
			return
		}

		atoken, rtoken, err := h.authService.GenerateTokens(u.ID, u.Role)
		if err != nil {
			delivery.HendleError(w, r, http.StatusInternalServerError, err)
			return
		}

		delivery.HendleRespond(w, r, http.StatusOK, response{
			AccessToken:  atoken,
			RefreshToken: rtoken,
		})
	}
}

func (h *AuthHendler) HandleTokensRefresh() http.HandlerFunc {
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

		token, err := h.authService.ParseToken(req.RefreshToken)
		if err != nil {
			delivery.HendleError(w, r, http.StatusUnauthorized, domain.ErrInvalidToken)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			tokenType, ok := claims["type"].(string)
			if !ok || tokenType != "refresh" {
				delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidToken)
				return
			}
			userID, ok := claims["sub"].(string)
			userid, err := strconv.ParseInt(userID, 10, 64)
			if err != nil {
				h.logger.Info(h.ctx, "failed to parse userID from token")
				delivery.HendleError(w, r, http.StatusUnauthorized, domain.ErrInvalidToken)
				return
			}
			if !ok {
				delivery.HendleError(w, r, http.StatusUnauthorized, domain.ErrInvalidToken)
				return
			}

			u, err := h.userService.UsersGetById(int(userid))
			if err != nil {
				delivery.HendleError(w, r, http.StatusNotFound, err)
				return
			}

			atoken, rtoken, err := h.authService.GenerateTokens(u.ID, u.Role)
			if err != nil {
				delivery.HendleError(w, r, http.StatusInternalServerError, err)
				return
			}

			delivery.HendleRespond(w, r, http.StatusOK, response{
				AccessToken:  atoken,
				RefreshToken: rtoken,
			})
		}
	}
}
