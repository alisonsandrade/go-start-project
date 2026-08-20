// Package handler prover a orquestração entre as rotas e os serviços.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/domain"
	"github.com/alisonsandrade/go-start-project/internal/middleware"
	"github.com/alisonsandrade/go-start-project/internal/repository"
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

// GetUser obtém os detalhes do perfil do usuário autenticado.
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
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserClaimsKey).(*token.CustomClaims)

	user, err := h.userService.GetUser(claims.UserID)
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

// UpdateUser atualiza os dados cadastrais do próprio usuário autenticado.
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
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserClaimsKey).(*token.CustomClaims)

	var dto domain.UpdateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "Corpo da requisição inválido")
		return
	}

	user, err := h.userService.UpdateUser(claims.UserID, dto)
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

// DeleteUser desativa/exclui logicamente a conta do usuário autenticado.
// @Summary      Deletar perfil do usuário logado (Soft Delete)
// @Description  Realiza a exclusão lógica do usuário autenticado no sistema
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  domain.MessageResponseDTO
// @Failure      401  {object}  domain.ErrorResponseDTO
// @Failure      404  {object}  domain.ErrorResponseDTO
// @Failure      500  {object}  domain.ErrorResponseDTO
// @Router       /api/users/me [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserClaimsKey).(*token.CustomClaims)

	err := h.userService.DeleteUser(claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}
		ErrorJSON(w, http.StatusInternalServerError, "erro interno ao deletar usuário")
		return
	}

	JSON(w, http.StatusOK, domain.MessageResponseDTO{
		Message: "Usuário desativado com sucesso",
	})
}

// Routes das rotas do Users
func (h *UserHandler) Routes(cfg *config.Config, roleRepo repository.RoleRepository) chi.Router {
	r := chi.NewRouter()

	// Protege todas as rotas da função com a exigência de token
	r.Use(middleware.AuthMiddleware(cfg))

	r.Get("/me", h.GetUser)
	r.Put("/me", h.UpdateUser)
	r.Delete("/me", h.DeleteUser)

	// Permission-gated subgroup. Module RBAC (ADMIN)
	r.Group(func(admin chi.Router) {
		admin.Use(middleware.RequirePermission(
			roleRepo,
			domain.PermissionListUsers,
		))
		admin.Get("/", h.ListUsers)
	})

	return r
}
