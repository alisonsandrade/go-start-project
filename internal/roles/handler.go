// Package roles handler provides orchestration between routes and services.
package roles

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/platform"
	"github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/pkg/apiresponse"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RoleHandler struct {
	roleService RoleService
}

func NewRoleHandler(roleService RoleService) *RoleHandler {
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
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/roles [get]
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.roleService.List()
	if err != nil {
		platform.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	platform.JSON(w, http.StatusOK, roles)
}

// GetRoleByID returns a role by ID.
// @Summary      Get role
// @Description  Returns a role by its UUID
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Role ID"
// @Success      200  {object}  domain.RoleEntity
// @Failure      400  {object}  apiresponse.ErrorResponse
// @Failure      404  {object}  apiresponse.ErrorResponse
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/roles/{id} [get]
func (h *RoleHandler) GetRoleByID(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid role id")
		return
	}

	role, err := h.roleService.GetByID(roleID)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}

		platform.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	platform.JSON(w, http.StatusOK, role)
}

// CreateRole creates a new role.
// @Summary      Create role
// @Description  Creates a new role
// @Tags         Roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        payload body domain.CreateRoleRequest true "Role payload"
// @Success      201  {object}  domain.RoleEntity
// @Failure      400  {object}  apiresponse.ErrorResponse
// @Failure      409  {object}  apiresponse.ErrorResponse
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/roles [post]
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	role := &domain.RoleEntity{
		Name:        req.Name,
		Description: req.Description,
	}

	createdRole, err := h.roleService.Create(role)
	if err != nil {
		if errors.Is(err, ErrRoleAlreadyExists) {
			platform.ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}

		platform.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	platform.JSON(w, http.StatusCreated, createdRole)
}

// UpdateRole updates an existing role.
// @Summary      Update role
// @Description  Updates an existing role
// @Tags         Roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Role ID"
// @Param        payload body domain.UpdateRoleRequest true "Role payload"
// @Success      200  {object}  domain.RoleEntity
// @Failure      400  {object}  apiresponse.ErrorResponse
// @Failure      404  {object}  apiresponse.ErrorResponse
// @Failure      409  {object}  apiresponse.ErrorResponse
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/roles/{id} [put]
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid role id")
		return
	}

	var req domain.UpdateRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	role := &domain.RoleEntity{
		ID:          roleID,
		Name:        req.Name,
		Description: req.Description,
	}

	updatedRole, err := h.roleService.Update(role)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, ErrSystemRoleImmutable) {
			platform.ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}

		platform.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	platform.JSON(w, http.StatusOK, updatedRole)
}

// DeleteRole deletes a role.
// @Summary      Delete role
// @Description  Deletes a role
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Role ID"
// @Success      200  {object}  apiresponse.MessageResponse
// @Failure      400  {object}  apiresponse.ErrorResponse
// @Failure      404  {object}  apiresponse.ErrorResponse
// @Failure      409  {object}  apiresponse.ErrorResponse
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/roles/{id} [delete]
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid role id")
		return
	}

	err = h.roleService.Delete(roleID)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, ErrSystemRoleImmutable) {
			platform.ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}

		platform.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	platform.JSON(w, http.StatusOK, apiresponse.MessageResponse{
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
// @Param        payload body domain.ReplacePermissionsRequest true "Permission IDs"
// @Success      200  {object}  apiresponse.MessageResponse
// @Failure      400  {object}  apiresponse.ErrorResponse
// @Failure      404  {object}  apiresponse.ErrorResponse
// @Failure      409  {object}  apiresponse.ErrorResponse
// @Failure      500  {object}  apiresponse.ErrorResponse
// @Router       /api/roles/{id}/permissions [put]
func (h *RoleHandler) ReplacePermissions(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid role id")
		return
	}

	var req domain.ReplacePermissionsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.roleService.ReplacePermissions(
		roleID,
		req.PermissionIDs,
	)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, ErrInvalidPermissions) {
			platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}

		if errors.Is(err, ErrSystemRoleImmutable) {
			platform.ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}

		platform.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	platform.JSON(w, http.StatusOK, apiresponse.MessageResponse{
		Message: "Role permissions updated successfully",
	})
}

// Routes registers role routes.
func (h *RoleHandler) Routes(cfg *config.Config, roleRepo RoleRepository) chi.Router {
	r := chi.NewRouter()

	r.Use(auth.AuthMiddleware(cfg))

	r.Group(func(admin chi.Router) {
		admin.Use(
			RequirePermission(roleRepo, domain.PermissionReadRole),
		)

		admin.Get("/", h.ListRoles)
		admin.Get("/{id}", h.GetRoleByID)

		admin.Post("/", h.CreateRole)
		admin.Put("/{id}", h.UpdateRole)
		admin.Delete("/{id}", h.DeleteRole)

		admin.Put("/{id}/permissions", h.ReplacePermissions)
	})

	return r
}
