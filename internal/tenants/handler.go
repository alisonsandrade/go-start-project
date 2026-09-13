// Package tenants handler provides tenant registration and lookup endpoints.
package tenants

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/platform"
	"github.com/alisonsandrade/go-start-project/internal/roles"
	rolesDomain "github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/internal/tenants/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Create registers a new company or institution tenant.
// @Summary      Register tenant
// @Description  Creates a tenant with its identification document and settings.
// @Tags         Tenants
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        payload body domain.CreateTenantRequest true "Tenant data"
// @Success      201  {object} domain.Tenant
// @Failure      400  {object} apiresponse.ErrorResponse
// @Failure      409  {object} apiresponse.ErrorResponse
// @Failure      500  {object} apiresponse.ErrorResponse
// @Router       /api/tenants [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request domain.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenant, err := h.service.Create(r.Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNameRequired), errors.Is(err, domain.ErrDocumentRequired), errors.Is(err, domain.ErrDocumentInvalid), errors.Is(err, domain.ErrSettingsInvalid):
			platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrAlreadyExists):
			platform.ErrorJSON(w, http.StatusConflict, err.Error())
		default:
			platform.ErrorJSON(w, http.StatusInternalServerError, "failed to create tenant")
		}
		return
	}
	platform.JSON(w, http.StatusCreated, tenant)
}

// GetByID returns a tenant by its UUID.
// @Summary      Get tenant
// @Description  Returns a tenant by its UUID.
// @Tags         Tenants
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Tenant ID"
// @Success      200  {object} domain.Tenant
// @Failure      400  {object} apiresponse.ErrorResponse
// @Failure      404  {object} apiresponse.ErrorResponse
// @Failure      500  {object} apiresponse.ErrorResponse
// @Router       /api/tenants/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	tenant, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}
		platform.ErrorJSON(w, http.StatusInternalServerError, "failed to find tenant")
		return
	}
	platform.JSON(w, http.StatusOK, tenant)
}

// GetCurrent returns the tenant from the authenticated user's claims.
// @Summary      Get current tenant
// @Description  Returns only the tenant associated with the authenticated user.
// @Tags         Tenants
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object} domain.Tenant
// @Failure      401  {object} apiresponse.ErrorResponse
// @Failure      404  {object} apiresponse.ErrorResponse
// @Failure      500  {object} apiresponse.ErrorResponse
// @Router       /api/tenants/me [get]
func (h *Handler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.UserClaimsKey).(*token.CustomClaims)
	if !ok || claims == nil || claims.TenantID == uuid.Nil {
		platform.ErrorJSON(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	tenant, err := h.service.FindByID(r.Context(), claims.TenantID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}
		platform.ErrorJSON(w, http.StatusInternalServerError, "failed to find tenant")
		return
	}
	platform.JSON(w, http.StatusOK, tenant)
}

// List returns a paginated list of tenants.
// @Summary      List tenants
// @Description  Returns tenants using page and limit pagination.
// @Tags         Tenants
// @Security     BearerAuth
// @Produce      json
// @Param        page  query    int  false  "Page number"
// @Param        limit query    int  false  "Items per page (default: 10, max: 100)"
// @Success      200  {object} domain.TenantPageResponse
// @Failure      401  {object} apiresponse.ErrorResponse
// @Failure      500  {object} apiresponse.ErrorResponse
// @Router       /api/tenants [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	params := pagination.ExtractParams(r, 10, 100)
	result, err := h.service.List(r.Context(), params)
	if err != nil {
		platform.ErrorJSON(w, http.StatusInternalServerError, "failed to list tenants")
		return
	}
	platform.JSON(w, http.StatusOK, result)
}

// Update changes a tenant's name, document, or settings.
// @Summary      Update tenant
// @Description  Updates tenant registration data.
// @Tags         Tenants
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Tenant ID"
// @Param        payload body domain.UpdateTenantRequest true "Tenant data"
// @Success      200  {object} domain.Tenant
// @Failure      400  {object} apiresponse.ErrorResponse
// @Failure      401  {object} apiresponse.ErrorResponse
// @Failure      404  {object} apiresponse.ErrorResponse
// @Failure      409  {object} apiresponse.ErrorResponse
// @Failure      500  {object} apiresponse.ErrorResponse
// @Router       /api/tenants/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	var request domain.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenant, err := h.service.Update(r.Context(), id, request)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrAlreadyExists):
			platform.ErrorJSON(w, http.StatusConflict, err.Error())
		case errors.Is(err, domain.ErrNameRequired), errors.Is(err, domain.ErrDocumentRequired), errors.Is(err, domain.ErrDocumentInvalid), errors.Is(err, domain.ErrSettingsInvalid):
			platform.ErrorJSON(w, http.StatusBadRequest, err.Error())
		default:
			platform.ErrorJSON(w, http.StatusInternalServerError, "failed to update tenant")
		}
		return
	}
	platform.JSON(w, http.StatusOK, tenant)
}

// Delete removes a tenant from the system.
// @Summary      Delete tenant
// @Description  Soft deletes a tenant. Requires tenant management permission.
// @Tags         Tenants
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Tenant ID"
// @Success      200  {object} apiresponse.MessageResponse
// @Failure      401  {object} apiresponse.ErrorResponse
// @Failure      403  {object} apiresponse.ErrorResponse
// @Failure      404  {object} apiresponse.ErrorResponse
// @Failure      500  {object} apiresponse.ErrorResponse
// @Router       /api/tenants/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		platform.ErrorJSON(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			platform.ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}
		platform.ErrorJSON(w, http.StatusInternalServerError, "failed to delete tenant")
		return
	}
	platform.JSON(w, http.StatusOK, map[string]string{"message": "tenant deleted successfully"})
}

func (h *Handler) Routes(cfg *config.Config, roleRepo roles.RoleRepository) chi.Router {
	r := chi.NewRouter()
	r.Group(func(protected chi.Router) {
		protected.Use(auth.AuthMiddleware(cfg))
		protected.Get("/me", h.GetCurrent)
		protected.Group(func(admin chi.Router) {
			admin.Use(roles.RequirePermission(roleRepo, rolesDomain.PermissionManageTenant))
			admin.Post("/", h.Create)
			admin.Get("/", h.List)
			admin.Get("/{id}", h.GetByID)
			admin.Put("/{id}", h.Update)
			admin.Delete("/{id}", h.Delete)
		})
	})
	return r
}
