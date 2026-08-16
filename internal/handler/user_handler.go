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

// Register processa a requisição para cadastrar um novo usuário comum.
// @Summary      Cadastro público de usuário (Role USER)
// @Description  Cadastra um novo usuário padrão no sistema
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        payload body domain.CreateUserDTO true "Dados do usuário"
// @Success      201  {object}  domain.User
// @Failure      400  {object}  domain.ErrorResponseDTO
// @Failure      500  {object}  domain.ErrorResponseDTO
// @Router       /api/users/register [post]
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
		ErrorJSON(w, http.StatusInternalServerError, "Erro interno ao processar a requisição")
		return
	}

	JSON(w, http.StatusCreated, user)
}

// Login autentica as credenciais do usuário e retorna o token JWT.
// @Summary      Autenticação de usuário
// @Description  Autentica credenciais e retorna o token JWT
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload body domain.LoginDTO true "Credenciais de login"
// @Success      200  {object}  domain.AuthResponseDTO
// @Failure      400  {object}  domain.ErrorResponseDTO
// @Failure      401  {object}  domain.ErrorResponseDTO
// @Failure      500  {object}  domain.ErrorResponseDTO
// @Router       /api/users/login [post]
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

// GetProfile obtém os detalhes do perfil do usuário autenticado.
// @Summary      Obter perfil do usuário autenticado
// @Description  Retorna os detalhes do usuário logado via Bearer JWT
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  domain.User
// @Failure      401  {object}  domain.ErrorResponseDTO
// @Failure      404  {object}  domain.ErrorResponseDTO
// @Failure      500  {object}  domain.ErrorResponseDTO
// @Router       /api/users/me [get]
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

// ListUsers retorna uma lista de todos os usuários cadastrados no sistema.
// @Summary      Listar todos os usuários (RBAC - ADMIN)
// @Description  Retorna lista de todos os usuários cadastrados. Acesso restrito a ADMIN.
// @Tags         Admin
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   domain.User
// @Failure      401  {object}  domain.ErrorResponseDTO
// @Failure      403  {object}  domain.ErrorResponseDTO
// @Failure      500  {object}  domain.ErrorResponseDTO
// @Router       /api/users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.ListUsers()
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, users)
}

// UpdateProfile atualiza os dados cadastrais do próprio usuário autenticado.
// @Summary      Atualizar perfil do usuário logado
// @Description  Atualiza dados cadastrais do próprio usuário autenticado
// @Tags         Users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        payload body domain.UpdateUserDTO true "Campos para atualização"
// @Success      200  {object}  domain.User
// @Failure      400  {object}  domain.ErrorResponseDTO
// @Failure      401  {object}  domain.ErrorResponseDTO
// @Failure      500  {object}  domain.ErrorResponseDTO
// @Router       /api/users/me [put]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserClaimsKey).(*token.CustomClaims)

	var dto domain.UpdateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "Corpo da requisição inválido")
		return
	}

	user, err := h.userService.UpdateProfile(claims.UserID, dto)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}
		ErrorJSON(w, http.StatusInternalServerError, "erro interno ao tentar atualizar o registro")
		return
	}

	JSON(w, http.StatusOK, user)
}

// Routes é a função de modularização do roteamento do user_handler
func (h *UserHandler) Routes(cfg *config.Config) chi.Router {
	r := chi.NewRouter()

	// Rotas públicas
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)

	// Rotas protegidas
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

		r.Get("/me", h.GetProfile)
		r.Put("/me", h.UpdateProfile)

		// Sub-grupo de rotas. Módulo RBAC. Apenas ADMIN
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(domain.RoleAdmim))
			r.Get("/", h.ListUsers)
		})
	})

	return r
}
