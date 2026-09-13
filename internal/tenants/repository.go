package tenants

import (
	"context"
	"errors"

	"github.com/alisonsandrade/go-start-project/internal/tenants/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	Update(ctx context.Context, tenant *domain.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	FindByDocument(ctx context.Context, document string) (*domain.Tenant, error)
	List(ctx context.Context, limit, offset int) ([]domain.Tenant, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, tenant *domain.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *repository) Update(ctx context.Context, tenant *domain.Tenant) error {
	return r.db.WithContext(ctx).
		Model(tenant).
		Select("name", "document", "settings").
		Updates(tenant).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Tenant{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	var tenant domain.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *repository) FindByDocument(ctx context.Context, document string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	if err := r.db.WithContext(ctx).Where("document = ?", document).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]domain.Tenant, int64, error) {
	var tenants []domain.Tenant
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Tenant{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tenants).Error
	return tenants, total, err
}
