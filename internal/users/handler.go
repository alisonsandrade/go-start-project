// Package users handler prover a orquestração entre as rotas e os serviços.
package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/platform"
	"github.com/alisonsandrade/go-start-project/internal/roles"
	rolesDomain "github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/internal/users/domain"
	pkgDomain "github.com/alisonsandrade/go-start-project/pkg/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService UserService
	validate    *validator.Validate
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

// GetUser obtém os detalhes do perfil do usuário autenticado.
// @Summary      Obter perfil do usuário autenticado
// @Description  Retorna os detalhes do usuário logado via Bearer JWT
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  domain.User
// @Failure      401  {object}  platform.ErrorResponse
// @Failure      404  {object}  platform.ErrorResponse
// @Failure      500  {object}  platform.ErrorResponse
// @Router       /api/users/me [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.UserClaimsKey).(*token.CustomClaims)

	user, err := h.userService.GetUser(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			platform.ErrorWithCodeJSON(w, http.StatusNotFound, platform.CodeUserNotFound, err.Error())
			return
		}
		platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, err.Error())
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
// @Failure      401   {object} platform.ErrorResponse
// @Failure      403   {object} platform.ErrorResponse
// @Failure      500   {object} platform.ErrorResponse
// @Router       /api/users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	params := pagination.ExtractParams(r, 10, 100)

	result, err := h.userService.ListUsers(r.Context(), params)
	if err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, "erro ao listar usuários")
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
// @Failure      400  {object}  platform.ErrorResponse
// @Failure      401  {object}  platform.ErrorResponse
// @Failure      409  {object}  platform.ErrorResponse
// @Failure      500  {object}  platform.ErrorResponse
// @Router       /api/users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.UserClaimsKey).(*token.CustomClaims)
	if !ok || claims == nil {
		platform.ErrorWithCodeJSON(w, http.StatusUnauthorized, platform.CodeUnauthorized, "Não autorizado")
		return
	}

	var dto domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidRequestBody, err.Error())
		return
	}

	if err := h.validate.Struct(dto); err != nil {
		platform.ValidationErrorJSON(w, err)
		return
	}

	user, err := h.userService.CreateUserAsAdmin(r.Context(), dto)
	if err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			platform.ErrorWithCodeJSON(w, http.StatusForbidden, platform.CodeForbidden, "Acesso restrito para administradores")
		case errors.Is(err, ErrEmailAlreadyExists):
			platform.ErrorWithCodeJSON(w, http.StatusConflict, platform.CodeEmailAlreadyExists, err.Error())
		case errors.Is(err, ErrInvalidRole):
			platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidRole, err.Error())
		case errors.Is(err, pkgDomain.ErrInvalidEmail):
			platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidEmail, err.Error())
		default:
			platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, err.Error())
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
// @Failure      400  {object}  platform.ErrorResponse
// @Failure      401  {object}  platform.ErrorResponse
// @Failure      404  {object}  platform.ErrorResponse
// @Failure      500  {object}  platform.ErrorResponse
// @Router       /api/users/me [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.UserClaimsKey).(*token.CustomClaims)

	var dto domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidRequestBody, "Corpo da requisição inválido")
		return
	}

	if err := h.validate.Struct(dto); err != nil {
		platform.ValidationErrorJSON(w, err)
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), claims.UserID, dto)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			platform.ErrorWithCodeJSON(w, http.StatusNotFound, platform.CodeUserNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			platform.ErrorWithCodeJSON(w, http.StatusForbidden, platform.CodeForbidden, "Acesso restrito para administradores")
		case errors.Is(err, ErrInvalidRole):
			platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidRole, err.Error())
		case errors.Is(err, pkgDomain.ErrInvalidEmail):
			platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidEmail, err.Error())
		default:
			platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, err.Error())
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
// @Success      200  {object}  platform.MessageResponse
// @Failure      401  {object}  platform.ErrorResponse
// @Failure      404  {object}  platform.ErrorResponse
// @Failure      500  {object}  platform.ErrorResponse
// @Router       /api/users/me [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.UserClaimsKey).(*token.CustomClaims)

	err := h.userService.DeleteUser(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			platform.ErrorWithCodeJSON(w, http.StatusNotFound, platform.CodeUserNotFound, err.Error())
			return
		}
		platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, "erro interno ao deletar usuário")
		return
	}

	platform.JSON(w, http.StatusOK, platform.MessageResponse{
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
// @Failure      400  {object}  platform.ErrorResponse
// @Failure      401  {object}  platform.ErrorResponse
// @Failure      404  {object}  platform.ErrorResponse
// @Failure      500  {object}  platform.ErrorResponse
// @Router       /api/users/{id} [get]
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	idUUID, err := uuid.Parse(userID)
	if err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidUUID, err.Error())
		return
	}

	user, err := h.userService.GetUser(r.Context(), idUUID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			platform.ErrorWithCodeJSON(w, http.StatusNotFound, platform.CodeUserNotFound, err.Error())
			return
		}
		platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, err.Error())
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
// @Failure      400  {object}  platform.ErrorResponse
// @Failure      401  {object}  platform.ErrorResponse
// @Failure      404  {object}  platform.ErrorResponse
// @Failure      500  {object}  platform.ErrorResponse
// @Router       /api/users/{id} [put]
func (h *UserHandler) UpdateUserAsAdmin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userUUID, err := uuid.Parse(id)
	if err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidUUID, err.Error())
		return
	}

	var dto domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidRequestBody, "Corpo da requisição inválido")
		return
	}

	if err := h.validate.Struct(dto); err != nil {
		platform.ValidationErrorJSON(w, err)
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), userUUID, dto)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			platform.ErrorWithCodeJSON(w, http.StatusNotFound, platform.CodeUserNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			platform.ErrorWithCodeJSON(w, http.StatusForbidden, platform.CodeForbidden, "Acesso restrito para administradores")
		case errors.Is(err, ErrInvalidRole):
			platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidRole, err.Error())
		case errors.Is(err, pkgDomain.ErrInvalidEmail):
			platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidEmail, err.Error())
		default:
			platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, err.Error())
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
// @Success      200  {object}  platform.MessageResponse
// @Failure      400  {object}  platform.ErrorResponse
// @Failure      401  {object}  platform.ErrorResponse
// @Failure      404  {object}  platform.ErrorResponse
// @Failure      500  {object}  platform.ErrorResponse
// @Router       /api/users/{id} [delete]
func (h *UserHandler) SoftDeleteUserAsAdmin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userUUID, er := uuid.Parse(id)
	if er != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidUUID, er.Error())
		return
	}

	err := h.userService.SoftDeleteUserAsAdmin(r.Context(), userUUID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			platform.ErrorWithCodeJSON(w, http.StatusNotFound, platform.CodeUserNotFound, err.Error())
			return
		}
		platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, "erro interno ao deletar usuário")
		return
	}

	platform.JSON(w, http.StatusOK, platform.MessageResponse{
		Message: "Usuário desativado com sucesso",
	})
}

// ChangeUserTenant moves a user to another tenant and assigns a role there.
// @Summary      Alterar tenant do usuário
// @Description  Move um usuário para outro tenant e aplica uma role pertencente ao destino.
// @Tags         Admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Param        payload body domain.ChangeUserTenantRequest true "Tenant e role de destino"
// @Success      204
// @Failure      400 {object} platform.ErrorResponse
// @Failure      401 {object} platform.ErrorResponse
// @Failure      403 {object} platform.ErrorResponse
// @Failure      404 {object} platform.ErrorResponse
// @Router       /api/users/{id}/tenant [patch]
func (h *UserHandler) ChangeUserTenant(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidUUID, "id do usuário inválido")
		return
	}

	var dto domain.ChangeUserTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeInvalidRequestBody, "corpo da requisição inválido")
		return
	}

	if err := h.validate.Struct(dto); err != nil {
		platform.ValidationErrorJSON(w, err)
		return
	}

	if err := h.userService.ChangeUserTenant(r.Context(), id, dto); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			platform.ErrorWithCodeJSON(w, http.StatusNotFound, platform.CodeNotFound, "usuário, tenant ou role não encontrado")
		case strings.Contains(err.Error(), "obrigatórios"), strings.Contains(err.Error(), "já pertence"):
			platform.ErrorWithCodeJSON(w, http.StatusBadRequest, platform.CodeBadRequest, err.Error())
		default:
			platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, "falha ao alterar tenant do usuário")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
		admin.Group(func(editor chi.Router) {
			editor.Use(roles.RequirePermission(
				roleRepo,
				rolesDomain.PermissionUpdateUser,
			))
			editor.Patch("/{id}/tenant", h.ChangeUserTenant)
		})
	})

	return r
}
