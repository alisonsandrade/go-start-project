package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/domain"
	"github.com/alisonsandrade/go-start-project/internal/middleware"
	"github.com/alisonsandrade/go-start-project/internal/service"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/go-chi/chi/v5"
)

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler backed by the given AuthService
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles public user registration.
// @Summary      Public user registration
// @Description  Registers a new user with the default USER role and returns a token pair
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      domain.CreateUserDTO  true  "User registration data"
// @Success      201      {object}  domain.AuthResponseDTO
// @Failure      400      {object}  domain.ErrorResponseDTO
// @Failure      409      {object}  domain.ErrorResponseDTO
// @Failure      500      {object}  domain.ErrorResponseDTO
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var dto domain.CreateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := dto.Validate(); err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	authResponse, err := h.authService.Register(dto)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}
		ErrorJSON(w, http.StatusInternalServerError, "failed to process registration")
		return
	}

	JSON(w, http.StatusCreated, authResponse)
}

// Login authenticates a user and returns a new token pair.
// @Summary      User authentication
// @Description  Authenticates credentials and returns a JWT access token and a refresh token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      domain.LoginDTO  true  "Login credentials"
// @Success      200      {object}  domain.AuthResponseDTO
// @Failure      400      {object}  domain.ErrorResponseDTO
// @Failure      401      {object}  domain.ErrorResponseDTO
// @Failure      500      {object}  domain.ErrorResponseDTO
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var dto domain.LoginDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	authResponse, err := h.authService.Login(dto)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrUserInactive) {
			ErrorJSON(w, http.StatusUnauthorized, err.Error())
			return
		}
		ErrorJSON(w, http.StatusInternalServerError, "failed to process login")
		return
	}

	JSON(w, http.StatusOK, authResponse)
}

// RefreshToken rotates the refresh token and returns a new token pair.
// @Summary      Refresh session (tokens)
// @Description  Receives the current refresh token and returns a new token pair, rotating the session
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      domain.RefreshTokenDTO  true  "Refresh token"
// @Success      200      {object}  domain.AuthResponseDTO
// @Failure      400      {object}  domain.ErrorResponseDTO
// @Failure      401      {object}  domain.ErrorResponseDTO
// @Router       /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var dto domain.RefreshTokenDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if dto.RefreshToken == "" {
		ErrorJSON(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	authResponse, err := h.authService.RefreshSession(dto.RefreshToken)
	if err != nil {
		ErrorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	JSON(w, http.StatusOK, authResponse)
}

// Logout invalidates the authenticated user's session.
// @Summary      Logout
// @Description  Invalidates the session by revoking all of the user's refresh tokens
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  domain.MessageResponseDTO
// @Failure      401  {object}  domain.ErrorResponseDTO
// @Failure      500  {object}  domain.ErrorResponseDTO
// @Router       /api/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserClaimsKey).(*token.CustomClaims)
	if !ok || claims == nil {
		ErrorJSON(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	if err := h.authService.Logout(claims.UserID); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to logout")
		return
	}

	JSON(w, http.StatusOK, domain.MessageResponseDTO{
		Message: "logged out successfully",
	})
}

// AuthRoutes returns the router with the authentication endpoints.
func (h *AuthHandler) AuthRoutes(cfg *config.Config) chi.Router {
	r := chi.NewRouter()

	// Public routes
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.RefreshToken)

	// Protected route (requires a valid JWT)
	r.Group(func(protected chi.Router) {
		protected.Use(middleware.AuthMiddleware(cfg))
		protected.Post("/logout", h.Logout)
	})

	return r
}
