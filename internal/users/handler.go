// Package users handler prover a orquestração entre as rotas e os serviços.
package users

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/platform"
	"github.com/alisonsandrade/go-start-project/internal/roles"
	rolesDomain "github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/internal/users/domain"
	"github.com/alisonsandrade/go-start-project/pkg/apiresponse"
	pkgDomain "github.com/alisonsandrade/go-start-project/pkg/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetUser obtém os detalhes do perfil do usuário autenticado.
// @Summary      Obter perfil do usuário autenticado
// @Description  Retorna os detalhes do usuário logado via Bearer JWT
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  domain.User
// @Failure      401  {object}  apiresponse.ErrorResponse
// @Failure      404  {object}  apiresponse.ErrorResponse
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/users/me [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.UserClaimsKey).(*token.CustomClaims)

	user, err := h.userService.GetUser(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}
		platform.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	platform.JSON(w, http.StatusOK, user)
}

// ListUsers retorna uma lista paginada de todos os usuários cadastrados no sistema.
// @Summary      Listar usuários paginados (RBAC - ADMIN)
// @Description  Retorna lista paginada de usuários cadastrados. Acesso restrito a ADMIN.
// @Tags         Admin
// @Security     BearerAuth
// @Produce      json
// @Param        page  query    int  false  "Número da página (padrão: 1)"
// @Param        limit query    int  false  "Itens por página (padrão: 10, máx: 100)"
// @Success      200   {object} domain.UserPageResponse
// @Failure      401   {object} apiresponse.ErrorResponse
// @Failure      403   {object} apiresponse.ErrorResponse
// @Failure      500   {object} apiresponse.ErrorResponse
// @Router       /api/users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	params := pagination.ExtractParams(r, 10, 100)

	result, err := h.userService.ListUsers(r.Context(), params)
	if err != nil {
		platform.ErrorJSON(w, http.StatusInternalServerError, "erro ao listar usuários")
		return
	}

	platform.JSON(w, http.StatusOK, result)
}

// CreateUser create a new user as common user or admin.
// @Summary      Criar um novo perfil do usuário
// @Description  Cria um novo usuário para acesso ao sistema.
// @Tags         Users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        payload body domain.CreateUserRequest true "Campos para cadastrar novo usuário"
// @Success      201  {object}  domain.UserResponse
// @Failure      400  {object}  apiresponse.ErrorResponse
// @Failure      401  {object}  apiresponse.ErrorResponse
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.UserClaimsKey).(*token.CustomClaims)
	if !ok || claims == nil {
		platform.ErrorJSON(w, http.StatusUnauthorized, "Não autorizado")
		return
	}

	var dto domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userService.CreateUserAsAdmin(r.Context(), dto)
	if err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			platform.ErrorJSON(w, http.StatusForbidden, "Acesso restrito para administradores")
		case errors.Is(err, ErrEmailAlreadyExists):
			platform.ErrorJSON(w, http.StatusConflict, err.Error()) // 409 Conflict é o padrão REST para duplicidade
		case errors.Is(err, domain.ErrUserNameTooShort):
			platform.ErrorJSON(w, http.StatusBadRequest, "O nome do usuário deve conter pelo menos 2 letras.")
		case errors.Is(err, domain.ErrUserNameTooLong):
			platform.ErrorJSON(w, http.StatusBadRequest, "O nome do usuário é muito longo (máximo de 100 caracteres).")
		case errors.Is(err, domain.ErrUserRoleRequired):
			platform.ErrorJSON(w, http.StatusBadRequest, "É obrigatório selecionar um perfil (Role) para o usuário.")
		case errors.Is(err, domain.ErrUserPhoneTooLong):
			platform.ErrorJSON(w, http.StatusBadRequest, "O número de telefone inserido é inválido (muito longo).")
		case errors.Is(err, domain.ErrUserAvatarInvalid):
			platform.ErrorJSON(w, http.StatusBadRequest, "O link da foto de perfil não é uma URL válida.")
		case errors.Is(err, ErrInvalidRole):
			platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, pkgDomain.ErrInvalidEmail):
			platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		default:
			platform.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	platform.JSON(w, http.StatusCreated, user)
}

// UpdateUser atualiza os dados cadastrais do próprio usuário autenticado.
// @Summary      Atualizar perfil do usuário logado
// @Description  Atualiza dados cadastrais do próprio usuário autenticado
// @Tags         Users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        payload body domain.UpdateUserRequest true "Campos para atualização"
// @Success      200  {object}  domain.User
// @Failure      400  {object}  apiresponse.ErrorResponse
// @Failure      401  {object}  apiresponse.ErrorResponse
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/users/me [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.UserClaimsKey).(*token.CustomClaims)

	var dto domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "Corpo da requisição inválido")
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), claims.UserID, dto)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			platform.ErrorJSON(w, http.StatusForbidden, "Acesso restrito para administradores")
		case errors.Is(err, domain.ErrUserNameTooShort):
			platform.ErrorJSON(w, http.StatusBadRequest, "O nome do usuário deve conter pelo menos 2 letras.")
		case errors.Is(err, domain.ErrUserNameTooLong):
			platform.ErrorJSON(w, http.StatusBadRequest, "O nome do usuário é muito longo (máximo de 100 caracteres).")
		case errors.Is(err, domain.ErrUserRoleRequired):
			platform.ErrorJSON(w, http.StatusBadRequest, "É obrigatório selecionar um perfil (Role) para o usuário.")
		case errors.Is(err, domain.ErrUserPhoneTooLong):
			platform.ErrorJSON(w, http.StatusBadRequest, "O número de telefone inserido é inválido (muito longo).")
		case errors.Is(err, domain.ErrUserAvatarInvalid):
			platform.ErrorJSON(w, http.StatusBadRequest, "O link da foto de perfil não é uma URL válida.")
		case errors.Is(err, ErrInvalidRole):
			platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, pkgDomain.ErrInvalidEmail):
			platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		default:
			platform.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	platform.JSON(w, http.StatusOK, user)
}

// DeleteUser desativa/exclui logicamente a conta do usuário autenticado.
// @Summary      Deletar perfil do usuário logado (Soft Delete)
// @Description  Realiza a exclusão lógica do usuário autenticado no sistema
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  apiresponse.MessageResponse
// @Failure      401  {object}  apiresponse.ErrorResponse
// @Failure      404  {object}  apiresponse.ErrorResponse
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/users/me [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.UserClaimsKey).(*token.CustomClaims)

	err := h.userService.DeleteUser(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}
		platform.ErrorJSON(w, http.StatusInternalServerError, "erro interno ao deletar usuário")
		return
	}

	platform.JSON(w, http.StatusOK, apiresponse.MessageResponse{
		Message: "Usuário desativado com sucesso",
	})
}

/**************************************************************************************************
*										handler's user admin
**************************************************************************************************/

// GetUserByID obtém os detalhes do perfil do usuário cujo ID foi passado por parâmetro.
// @Summary      Obter perfil do usuário
// @Description  Retorna os detalhes do usuário por ID
// @Tags         Admin
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200  {object}  domain.User
// @Failure      401  {object}  apiresponse.ErrorResponse
// @Failure      404  {object}  apiresponse.ErrorResponse
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/users/{id} [get]
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	idUUID, err := uuid.Parse(userID)
	if err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userService.GetUser(r.Context(), idUUID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}
		platform.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	platform.JSON(w, http.StatusOK, user)
}

// UpdateUserAsAdmin udpate specific user's data by ID.
// @Summary      Update perfil user
// @Description  Update specific user's data by ID from admin user
// @Tags		 Admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Param        payload body domain.UpdateUserRequest true "Fields for udpate"
// @Success      200  {object}  domain.User
// @Failure      400  {object}  apiresponse.ErrorResponse
// @Failure      401  {object}  apiresponse.ErrorResponse
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/users/{id} [put]
func (h *UserHandler) UpdateUserAsAdmin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userUUID, err := uuid.Parse(id)
	if err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	var dto domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "Corpo da requisição inválido")
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), userUUID, dto)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			platform.ErrorJSON(w, http.StatusForbidden, "Acesso restrito para administradores")
		case errors.Is(err, domain.ErrUserNameTooShort):
			platform.ErrorJSON(w, http.StatusBadRequest, "O nome do usuário deve conter pelo menos 2 letras.")
		case errors.Is(err, domain.ErrUserNameTooLong):
			platform.ErrorJSON(w, http.StatusBadRequest, "O nome do usuário é muito longo (máximo de 100 caracteres).")
		case errors.Is(err, domain.ErrUserRoleRequired):
			platform.ErrorJSON(w, http.StatusBadRequest, "É obrigatório selecionar um perfil (Role) para o usuário.")
		case errors.Is(err, domain.ErrUserPhoneTooLong):
			platform.ErrorJSON(w, http.StatusBadRequest, "O número de telefone inserido é inválido (muito longo).")
		case errors.Is(err, domain.ErrUserAvatarInvalid):
			platform.ErrorJSON(w, http.StatusBadRequest, "O link da foto de perfil não é uma URL válida.")
		case errors.Is(err, ErrInvalidRole):
			platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, pkgDomain.ErrInvalidEmail):
			platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		default:
			platform.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	platform.JSON(w, http.StatusOK, user)
}

// SoftDeleteUserAsAdmin deactivate logical the accounts user by ID.
// @Summary      Soft delete perfil user by ID
// @Description  Deletes a logical user from the system
// @Tags         Admin
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200  {object}  apiresponse.MessageResponse
// @Failure      401  {object}  apiresponse.ErrorResponse
// @Failure      404  {object}  apiresponse.ErrorResponse
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/users/{id} [delete]
func (h *UserHandler) SoftDeleteUserAsAdmin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userUUID, er := uuid.Parse(id)
	if er != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, er.Error())
		return
	}

	err := h.userService.SoftDeleteUserAsAdmin(r.Context(), userUUID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}
		platform.ErrorJSON(w, http.StatusInternalServerError, "erro interno ao deletar usuário")
		return
	}

	platform.JSON(w, http.StatusOK, apiresponse.MessageResponse{
		Message: "Usuário desativado com sucesso",
	})
}

/**************************************************************************************************
*										handler's routers user
**************************************************************************************************/

// Routes das rotas do Users
func (h *UserHandler) Routes(cfg *config.Config, roleRepo roles.RoleRepository) chi.Router {
	r := chi.NewRouter()

	// Protege todas as rotas da função com a exigência de token
	r.Use(auth.AuthMiddleware(cfg))

	r.Post("/", h.CreateUser)
	r.Get("/me", h.GetUser)
	r.Put("/me", h.UpdateUser)
	r.Delete("/me", h.DeleteUser)

	// Permission-gated subgroup. Module RBAC (ADMIN)
	r.Group(func(admin chi.Router) {
		admin.Use(roles.RequirePermission(
			roleRepo,
			rolesDomain.PermissionListUsers,
		))
		admin.Get("/{id}", h.GetUserByID)
		admin.Put("/{id}", h.UpdateUserAsAdmin)
		admin.Delete("/{id}", h.SoftDeleteUserAsAdmin)
		admin.Get("/", h.ListUsers)
	})

	return r
}
