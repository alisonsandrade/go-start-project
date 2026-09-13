package tenants

import (
	"context"
	"errors"

	"github.com/alisonsandrade/go-start-project/internal/tenants/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

var ErrAlreadyExists = errors.New("tenant already exists")
var ErrNotFound = errors.New("tenant not found")

type Service interface {
	Create(ctx context.Context, request domain.CreateTenantRequest) (*domain.Tenant, error)
	Update(ctx context.Context, id uuid.UUID, request domain.UpdateTenantRequest) (*domain.Tenant, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	List(ctx context.Context, params pagination.Params) (pagination.PageResult[domain.Tenant], error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, request domain.CreateTenantRequest) (*domain.Tenant, error) {
	tenant := &domain.Tenant{
		Name:     request.Name,
		Document: request.Document,
		Settings: datatypes.JSON(request.Settings),
	}
	if err := tenant.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByDocument(ctx, tenant.Document)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyExists
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *service) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	tenant, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, ErrNotFound
	}
	return tenant, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, request domain.UpdateTenantRequest) (*domain.Tenant, error) {
	tenant, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	previousDocument := tenant.Document
	if request.Name != nil {
		tenant.Name = *request.Name
	}
	if request.Document != nil {
		tenant.Document = *request.Document
	}
	if request.Settings != nil {
		tenant.Settings = datatypes.JSON(request.Settings)
	}
	if err := tenant.Validate(); err != nil {
		return nil, err
	}

	if tenant.Document != previousDocument {
		duplicate, err := s.repo.FindByDocument(ctx, tenant.Document)
		if err != nil {
			return nil, err
		}
		if duplicate != nil && duplicate.ID != tenant.ID {
			return nil, ErrAlreadyExists
		}
	}

	if err := s.repo.Update(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *service) List(ctx context.Context, params pagination.Params) (pagination.PageResult[domain.Tenant], error) {
	tenants, total, err := s.repo.List(ctx, params.Limit, params.Offset())
	if err != nil {
		return pagination.PageResult[domain.Tenant]{}, err
	}
	return pagination.NewPageResult(tenants, total, params), nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
