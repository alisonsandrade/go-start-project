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

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var dto domain.CreateUserDTO

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	user, err := h.userService.Register(dto)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}
	}

	JSON(w, http.StatusCreated, user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var dto domain.LoginDTO

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	authResponse, err := h.userService.Login(dto)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			ErrorJSON(w, http.StatusUnauthorized, err.Error())
			return
		}
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, authResponse)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserClaimsKey).(*token.CustomClaims)

	user, err := h.userService.GetProfile(claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, user)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.ListUsers()
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, users)
}

// Routes é a função de modularização do roteamento do user_handler
func (h *UserHandler) Routes(cfg *config.Config) chi.Router {
	r := chi.NewRouter()

	// Rotas públicas
	r.Post("/login", h.Login)
	r.Post("/", h.Register)

	// Rotas protegidas
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

		r.Get("/me", h.GetProfile)

		// Sub-grupo de rotas. Módulo RBAC. Apenas ADMIN
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(domain.RoleAdmim))
			r.Get("/", h.ListUsers)
		})
	})

	return r
}
