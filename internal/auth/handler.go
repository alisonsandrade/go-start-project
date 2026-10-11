// Package auth
package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/auth/domain"
	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/platform"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
)

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	authService AuthService
	validate    *validator.Validate
}

// NewAuthHandler creates a new AuthHandler backed by the given AuthService
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
	}
}

// Register handles public user registration.
// @Summary      Public user registration
// @Description  Registers a new user with the default USER role and returns a token pair
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      domain.RegisterRequest  true  "User registration data"
// @Success      201      {object}  domain.AuthResponseDTO
// @Failure      400      {object}  platform.ErrorResponse
// @Failure      409      {object}  platform.ErrorResponse
// @Failure      422      {object}  platform.ErrorResponse
// @Failure      500      {object}  platform.ErrorResponse
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var dto domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidRequestBody, "Corpo da requisição inválido")
		return
	}

	dto.Name = strings.TrimSpace(dto.Name)
	dto.Email = strings.ToLower(strings.TrimSpace(dto.Email))
	dto.Phone = strings.TrimSpace(dto.Phone)
	dto.AvatarURL = strings.TrimSpace(dto.AvatarURL)
	dto.JobTitle = strings.TrimSpace(dto.JobTitle)
	dto.Bio = strings.TrimSpace(dto.Bio)

	if err := h.validate.Struct(dto); err != nil {
		platform.ValidationErrorJSON(w, err)
		return
	}

	authResponse, err := h.authService.Register(r.Context(), dto)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			platform.ErrorWithCodeJSON(w, http.StatusConflict, platform.CodeEmailAlreadyExists, err.Error())
			return
		}
		platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, "Falha ao processar o cadastro")
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
// @Failure      400      {object}  platform.ErrorResponse
// @Failure      401      {object}  platform.ErrorResponse
// @Failure      422      {object}  platform.ErrorResponse
// @Failure      500      {object}  platform.ErrorResponse
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var dto domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidRequestBody, "Corpo da requisição inválido")
		return
	}

	dto.Email = strings.ToLower(strings.TrimSpace(dto.Email))
	dto.Password = strings.TrimSpace(dto.Password)

	if err := h.validate.Struct(dto); err != nil {
		platform.ValidationErrorJSON(w, err)
		return
	}

	authResponse, err := h.authService.Login(r.Context(), dto)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrUserInactive) {
			platform.ErrorWithCodeJSON(w, http.StatusUnauthorized, platform.CodeInvalidCredentials, err.Error())
			return
		}
		platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, "Falha ao processar o login")
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
// @Failure      400      {object}  platform.ErrorResponse
// @Failure      401      {object}  platform.ErrorResponse
// @Failure      422      {object}  platform.ErrorResponse
// @Router       /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var dto domain.RefreshTokenDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidRequestBody, "Corpo da requisição inválido")
		return
	}

	dto.RefreshToken = strings.TrimSpace(dto.RefreshToken)

	if err := h.validate.Struct(dto); err != nil {
		platform.ValidationErrorJSON(w, err)
		return
	}

	authResponse, err := h.authService.RefreshSession(r.Context(), dto.RefreshToken)
	if err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusUnauthorized, platform.CodeUnauthorized, err.Error())
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
// @Success      200  {object}  platform.MessageResponse
// @Failure      401  {object}  platform.ErrorResponse
// @Failure      500  {object}  platform.ErrorResponse
// @Router       /api/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsKey).(*token.CustomClaims)
	if !ok || claims == nil {
		platform.ErrorWithCodeJSON(w, http.StatusUnauthorized, platform.CodeUnauthorized, "Não autenticado")
		return
	}

	if err := h.authService.Logout(r.Context(), claims.UserID); err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, "Falha ao efetuar logout")
		return
	}

	platform.JSON(w, http.StatusOK, platform.MessageResponse{
		Message: "Sessão encerrada com sucesso",
	})
}

// ForgotPassword solicita o link de recuperação de senha.
// @Summary      Request password reset
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      domain.ForgotPasswordDTO  true  "Email"
// @Success      200      {object}  platform.MessageResponse
// @Failure      400      {object}  platform.ErrorResponse
// @Failure      422      {object}  platform.ErrorResponse
// @Router       /api/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var dto domain.ForgotPasswordDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidRequestBody, "Corpo da requisição inválido")
		return
	}

	dto.Email = strings.ToLower(strings.TrimSpace(dto.Email))

	if err := h.validate.Struct(dto); err != nil {
		platform.ValidationErrorJSON(w, err)
		return
	}

	err := h.authService.ForgotPassword(r.Context(), dto.Email)
	if err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, "Falha ao processar solicitação")
		return
	}

	platform.JSON(w, http.StatusOK, platform.MessageResponse{
		Message: "Se o e-mail estiver cadastrado, você receberá um link de redefinição em breve.",
	})
}

// ResetPassword redefine a senha usando o token.
// @Summary      Reset password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      domain.ResetPasswordDTO  true  "Token and new password"
// @Success      200      {object}  platform.MessageResponse
// @Failure      400      {object}  platform.ErrorResponse
// @Failure      422      {object}  platform.ErrorResponse
// @Router       /api/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var dto domain.ResetPasswordDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidRequestBody, "Corpo da requisição inválido")
		return
	}

	dto.Token = strings.TrimSpace(dto.Token)
	dto.NewPassword = strings.TrimSpace(dto.NewPassword)

	if err := h.validate.Struct(dto); err != nil {
		platform.ValidationErrorJSON(w, err)
		return
	}

	err := h.authService.ResetPassword(r.Context(), dto.Token, dto.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, ErrResetTokenInvalid):
			platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidResetToken, err.Error())
		case errors.Is(err, ErrResetTokenExpired):
			platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeExpiredResetToken, err.Error())
		default:
			platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeBadRequest, err.Error())
		}
		return
	}

	platform.JSON(w, http.StatusOK, platform.MessageResponse{
		Message: "Senha redefinida com sucesso.",
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
// @Success      200      {object}  platform.MessageResponse
// @Failure      400      {object}  platform.ErrorResponse
// @Failure      401      {object}  platform.ErrorResponse
// @Failure      422      {object}  platform.ErrorResponse
// @Failure      500      {object}  platform.ErrorResponse
// @Router       /api/auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsKey).(*token.CustomClaims)
	if !ok || claims == nil {
		platform.ErrorWithCodeJSON(w, http.StatusUnauthorized, platform.CodeUnauthorized, "Não autenticado")
		return
	}

	var dto domain.ChangePasswordDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidRequestBody, "Corpo da requisição inválido")
		return
	}

	dto.CurrentPassword = strings.TrimSpace(dto.CurrentPassword)
	dto.NewPassword = strings.TrimSpace(dto.NewPassword)

	if err := h.validate.Struct(dto); err != nil {
		platform.ValidationErrorJSON(w, err)
		return
	}

	err := h.authService.ChangePassword(r.Context(), claims.UserID, dto)
	if err != nil {
		if errors.Is(err, ErrCurrentPasswordIncorrect) {
			platform.ErrorWithCodeJSON(w, http.StatusUnauthorized, platform.CodeIncorrectPassword, err.Error())
			return
		}
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeBadRequest, err.Error())
		return
	}

	platform.JSON(w, http.StatusOK, platform.MessageResponse{
		Message: "Senha alterada com sucesso",
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
	r.With(TenantContextMiddleware, registerRateLimit).Post("/register", h.Register)
	r.With(TenantContextMiddleware, authRateLimit).Post("/login", h.Login)
	r.With(TenantContextMiddleware).Post("/refresh", h.RefreshToken)
	r.With(TenantContextMiddleware).Post("/forgot-password", h.ForgotPassword)
	r.With(TenantContextMiddleware).Post("/reset-password", h.ResetPassword)

	// Protected route (requires a valid JWT)
	r.Group(func(protected chi.Router) {
		protected.Use(AuthMiddleware(cfg))
		protected.Post("/logout", h.Logout)
		protected.Post("/change-password", h.ChangePassword)
	})

	return r
}
