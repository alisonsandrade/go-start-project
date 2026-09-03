// Package auth
package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/auth/domain"
	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/platform"
	"github.com/alisonsandrade/go-start-project/pkg/apiresponse"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/go-chi/chi/v5"
)

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	authService AuthService
}

// NewAuthHandler creates a new AuthHandler backed by the given AuthService
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles public user registration.
// @Summary      Public user registration
// @Description  Registers a new user with the default USER role and returns a token pair
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      domain.RegisterRequest  true  "User registration data"
// @Success      201      {object}  domain.AuthResponseDTO
// @Failure      400      {object}  apiresponse.ErrorResponse
// @Failure      409      {object}  apiresponse.ErrorResponse
// @Failure      500      {object}  apiresponse.ErrorResponse
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var dto domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := dto.Validate(); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	authResponse, err := h.authService.Register(r.Context(), dto)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			platform.ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}
		platform.ErrorJSON(w, http.StatusInternalServerError, "failed to process registration")
		return
	}

	platform.JSON(w, http.StatusCreated, authResponse)
}

// Login authenticates a user and returns a new token pair.
// @Summary      User authentication
// @Description  Authenticates credentials and returns a JWT access token and a refresh token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      domain.LoginRequest  true  "Login credentials"
// @Success      200      {object}  domain.AuthResponseDTO
// @Failure      400      {object}  apiresponse.ErrorResponse
// @Failure      401      {object}  apiresponse.ErrorResponse
// @Failure      500      {object}  apiresponse.ErrorResponse
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var dto domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	authResponse, err := h.authService.Login(r.Context(), dto)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrUserInactive) {
			platform.ErrorJSON(w, http.StatusUnauthorized, err.Error())
			return
		}
		platform.ErrorJSON(w, http.StatusInternalServerError, "failed to process login")
		return
	}

	platform.JSON(w, http.StatusOK, authResponse)
}

// RefreshToken rotates the refresh token and returns a new token pair.
// @Summary      Refresh session (tokens)
// @Description  Receives the current refresh token and returns a new token pair, rotating the session
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      domain.RefreshTokenDTO  true  "Refresh token"
// @Success      200      {object}  domain.AuthResponseDTO
// @Failure      400      {object}  apiresponse.ErrorResponse
// @Failure      401      {object}  apiresponse.ErrorResponse
// @Router       /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var dto domain.RefreshTokenDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if dto.RefreshToken == "" {
		platform.ErrorJSON(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	authResponse, err := h.authService.RefreshSession(r.Context(), dto.RefreshToken)
	if err != nil {
		platform.ErrorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	platform.JSON(w, http.StatusOK, authResponse)
}

// Logout invalidates the authenticated user's session.
// @Summary      Logout
// @Description  Invalidates the session by revoking all of the user's refresh tokens
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  apiresponse.MessageResponse
// @Failure      401  {object}  apiresponse.ErrorResponse
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsKey).(*token.CustomClaims)
	if !ok || claims == nil {
		platform.ErrorJSON(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	if err := h.authService.Logout(r.Context(), claims.UserID); err != nil {
		platform.ErrorJSON(w, http.StatusInternalServerError, "failed to logout")
		return
	}

	platform.JSON(w, http.StatusOK, apiresponse.MessageResponse{
		Message: "logged out successfully",
	})
}

// ForgotPassword solicita o link de recuperação de senha.
// @Summary      Request password reset
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      domain.ForgotPasswordDTO  true  "Email"
// @Success      200      {object}  apiresponse.MessageResponse
// @Failure      400      {object}  apiresponse.ErrorResponse
// @Router       /api/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var dto domain.ForgotPasswordDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.authService.ForgotPassword(r.Context(), dto.Email)
	if err != nil {
		platform.ErrorJSON(w, http.StatusInternalServerError, "failed to process request")
		return
	}

	platform.JSON(w, http.StatusOK, apiresponse.MessageResponse{
		Message: "If the email is registered, you will receive a reset link shortly.",
	})
}

// ResetPassword redefine a senha usando o token.
// @Summary      Reset password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      domain.ResetPasswordDTO  true  "Token and new password"
// @Success      200      {object}  apiresponse.MessageResponse
// @Failure      400      {object}  apiresponse.ErrorResponse
// @Router       /api/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var dto domain.ResetPasswordDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.authService.ResetPassword(r.Context(), dto.Token, dto.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, ErrResetTokenInvalid):
			platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrResetTokenExpired):
			platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		// Captura regras do pacote de domínio (ex: tamanho mínimo da senha)
		default:
			// Se o err vier do pkgDomain.NewPassword, ele cai aqui com status 400
			platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	platform.JSON(w, http.StatusOK, apiresponse.MessageResponse{
		Message: "Password has been reset successfully.",
	})
}

// ChangePassword altera a senha do usuário autenticado.
// @Summary      Change password
// @Description  Altera a senha do usuário autenticado exigindo a confirmação da senha atual
// @Tags         Auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        payload  body      domain.ChangePasswordDTO  true  "Current and new password"
// @Success      200      {object}  apiresponse.MessageResponse
// @Failure      400      {object}  apiresponse.ErrorResponse
// @Failure      401      {object}  apiresponse.ErrorResponse
// @Failure      500      {object}  apiresponse.ErrorResponse
// @Router       /api/auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// 1. Extrai as claims injetadas pelo AuthMiddleware
	claims, ok := r.Context().Value(UserClaimsKey).(*token.CustomClaims)
	if !ok || claims == nil {
		platform.ErrorJSON(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var dto domain.ChangePasswordDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 2. Executa a troca
	err := h.authService.ChangePassword(r.Context(), claims.UserID, dto)
	if err != nil {
		if errors.Is(err, ErrCurrentPasswordIncorrect) {
			platform.ErrorJSON(w, http.StatusUnauthorized, err.Error())
			return
		}
		// Trata erros de validação da nova senha (ex: fraca ou curta)
		platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	platform.JSON(w, http.StatusOK, apiresponse.MessageResponse{
		Message: "password changed successfully",
	})
}

// AuthRoutes returns the router with the authentication endpoints.
func (h *AuthHandler) AuthRoutes(cfg *config.Config) chi.Router {
	r := chi.NewRouter()

	// Strict rate limiter for login/reset: 5 requests per minute per IP
	authRateLimit := platform.RateLimiter(5, time.Minute)

	// Moderate rate limiter for registration: 10 requests per minute per IP
	registerRateLimit := platform.RateLimiter(10, time.Minute)

	// Public routes
	r.With(registerRateLimit).Post("/register", h.Register)
	r.With(authRateLimit).Post("/login", h.Login)
	r.Post("/refresh", h.RefreshToken)
	r.Post("/forgot-password", h.ForgotPassword)
	r.Post("/reset-password", h.ResetPassword)

	// Protected route (requires a valid JWT)
	r.Group(func(protected chi.Router) {
		protected.Use(AuthMiddleware(cfg))
		protected.Post("/logout", h.Logout)
		protected.Post("/change-password", h.ChangePassword)
	})

	return r
}
