// Package handler provides orchestration between routes and services.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/domain"
	"github.com/alisonsandrade/go-start-project/internal/dto"
	"github.com/alisonsandrade/go-start-project/internal/middleware"
	"github.com/alisonsandrade/go-start-project/internal/repository"
	"github.com/alisonsandrade/go-start-project/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RoleHandler struct {
	roleService service.RoleService
}

func NewRoleHandler(roleService service.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// ListRoles lists all roles.
// @Summary      List roles
// @Description  Returns all roles registered in the system
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   domain.RoleEntity
// @Failure      500  {object}  domain.ErrorResponseDTO
// @Router       /api/roles [get]
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.roleService.List()
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, roles)
}

// GetRoleByID returns a role by ID.
// @Summary      Get role
// @Description  Returns a role by its UUID
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Role ID"
// @Success      200  {object}  domain.RoleEntity
// @Failure      400  {object}  domain.ErrorResponseDTO
// @Failure      404  {object}  domain.ErrorResponseDTO
// @Failure      500  {object}  domain.ErrorResponseDTO
// @Router       /api/roles/{id} [get]
func (h *RoleHandler) GetRoleByID(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid role id")
		return
	}

	role, err := h.roleService.GetByID(roleID)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}

		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, role)
}

// CreateRole creates a new role.
// @Summary      Create role
// @Description  Creates a new role
// @Tags         Roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        payload body dto.CreateRoleRequest true "Role payload"
// @Success      201  {object}  domain.RoleEntity
// @Failure      400  {object}  domain.ErrorResponseDTO
// @Failure      409  {object}  domain.ErrorResponseDTO
// @Failure      500  {object}  domain.ErrorResponseDTO
// @Router       /api/roles [post]
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	role := &domain.RoleEntity{
		Name:        req.Name,
		Description: req.Description,
	}

	createdRole, err := h.roleService.Create(role)
	if err != nil {
		if errors.Is(err, service.ErrRoleAlreadyExists) {
			ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}

		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusCreated, createdRole)
}

// UpdateRole updates an existing role.
// @Summary      Update role
// @Description  Updates an existing role
// @Tags         Roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Role ID"
// @Param        payload body dto.UpdateRoleRequest true "Role payload"
// @Success      200  {object}  domain.RoleEntity
// @Failure      400  {object}  domain.ErrorResponseDTO
// @Failure      404  {object}  domain.ErrorResponseDTO
// @Failure      409  {object}  domain.ErrorResponseDTO
// @Failure      500  {object}  domain.ErrorResponseDTO
// @Router       /api/roles/{id} [put]
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid role id")
		return
	}

	var req dto.UpdateRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	role := &domain.RoleEntity{
		ID:          roleID,
		Name:        req.Name,
		Description: req.Description,
	}

	updatedRole, err := h.roleService.Update(role)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, service.ErrSystemRoleImmutable) {
			ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}

		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, updatedRole)
}

// DeleteRole deletes a role.
// @Summary      Delete role
// @Description  Deletes a role
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Role ID"
// @Success      200  {object}  domain.MessageResponseDTO
// @Failure      400  {object}  domain.ErrorResponseDTO
// @Failure      404  {object}  domain.ErrorResponseDTO
// @Failure      409  {object}  domain.ErrorResponseDTO
// @Failure      500  {object}  domain.ErrorResponseDTO
// @Router       /api/roles/{id} [delete]
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid role id")
		return
	}

	err = h.roleService.Delete(roleID)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, service.ErrSystemRoleImmutable) {
			ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}

		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, domain.MessageResponseDTO{
		Message: "Role deleted successfully",
	})
}

// ReplacePermissions replaces all permissions assigned to a role.
// @Summary      Replace role permissions
// @Description  Replaces all permissions assigned to a role
// @Tags         Roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Role ID"
// @Param        payload body dto.ReplacePermissionsRequest true "Permission IDs"
// @Success      200  {object}  domain.MessageResponseDTO
// @Failure      400  {object}  domain.ErrorResponseDTO
// @Failure      404  {object}  domain.ErrorResponseDTO
// @Failure      409  {object}  domain.ErrorResponseDTO
// @Failure      500  {object}  domain.ErrorResponseDTO
// @Router       /api/roles/{id}/permissions [put]
func (h *RoleHandler) ReplacePermissions(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid role id")
		return
	}

	var req dto.ReplacePermissionsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.roleService.ReplacePermissions(
		roleID,
		req.PermissionIDs,
	)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, service.ErrInvalidPermissions) {
			ErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}

		if errors.Is(err, service.ErrSystemRoleImmutable) {
			ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}

		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, domain.MessageResponseDTO{
		Message: "Role permissions updated successfully",
	})
}

// Routes registers role routes.
func (h *RoleHandler) Routes(cfg *config.Config, roleRepo repository.RoleRepository) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.AuthMiddleware(cfg))

	r.Group(func(admin chi.Router) {
		admin.Use(middleware.RequireRole(
			domain.RoleAdmin,
		))

		admin.Get("/", h.ListRoles)
		admin.Get("/{id}", h.GetRoleByID)

		admin.Post("/", h.CreateRole)
		admin.Put("/{id}", h.UpdateRole)
		admin.Delete("/{id}", h.DeleteRole)

		admin.Put("/{id}/permissions", h.ReplacePermissions)
	})

	return r
}
