package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/internal/auth/domain"
	baseDomain "github.com/alisonsandrade/go-start-project/internal/domain"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupTokenTestDB initializes a compatible in-memory SQLite database.
func setupTokenTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE refresh_tokens (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			updated_at DATETIME,
			deleted_at DATETIME,
			user_id TEXT NOT NULL,
			token TEXT NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			created_at DATETIME
		);
	`).Error
	assert.NoError(t, err)

	return db
}

func TestTokenRepository_AllOperations(t *testing.T) {
	db := setupTokenTestDB(t)
	repo := auth.NewTokenRepository(db)
	ctx := context.WithValue(context.Background(), token.ClaimsContextKey, &token.CustomClaims{TenantID: token.DefaultTenantID})

	userID1 := uuid.New()
	userID2 := uuid.New()

	token1 := &domain.RefreshToken{
		BaseModelTenant: baseDomain.BaseModelTenant{BaseModel: baseDomain.BaseModel{ID: uuid.New(), CreatedAt: time.Now()}},
		UserID:          userID1,
		Token:           "token-abc-123",
		ExpiresAt:       time.Now().Add(24 * time.Hour),
	}

	token2 := &domain.RefreshToken{
		BaseModelTenant: baseDomain.BaseModelTenant{BaseModel: baseDomain.BaseModel{ID: uuid.New(), CreatedAt: time.Now()}},
		UserID:          userID1,
		Token:           "token-def-456",
		ExpiresAt:       time.Now().Add(24 * time.Hour),
	}

	token3 := &domain.RefreshToken{
		BaseModelTenant: baseDomain.BaseModelTenant{BaseModel: baseDomain.BaseModel{ID: uuid.New(), CreatedAt: time.Now()}},
		UserID:          userID2,
		Token:           "token-ghi-789",
		ExpiresAt:       time.Now().Add(24 * time.Hour),
	}

	t.Run("Create and FindByToken succeed", func(t *testing.T) {
		err := repo.Create(ctx, token1)
		assert.NoError(t, err)

		found, err := repo.FindByToken(ctx, token1.Token)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, token1.ID, found.ID)
		assert.Equal(t, token1.UserID, found.UserID)
		assert.Equal(t, token1.Token, found.Token)
	})

	t.Run("FindByToken returns an error when token does not exist", func(t *testing.T) {
		found, err := repo.FindByToken(ctx, "token-inexistente")
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("Delete removes a specific token", func(t *testing.T) {
		err := repo.Create(ctx, token3)
		assert.NoError(t, err)

		// Delete token3.
		err = repo.Delete(ctx, token3.Token)
		assert.NoError(t, err)

		// Confirm it no longer exists.
		found, err := repo.FindByToken(ctx, token3.Token)
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("DeleteByUserID revokes all tokens for a specific user", func(t *testing.T) {
		// Insert the second token for userID1.
		err := repo.Create(ctx, token2)
		assert.NoError(t, err)

		// Delete all tokens associated with userID1 (token1 and token2).
		err = repo.DeleteByUserID(ctx, userID1)
		assert.NoError(t, err)

		// Confirm neither token can be found.
		_, err1 := repo.FindByToken(ctx, token1.Token)
		assert.Error(t, err1)

		_, err2 := repo.FindByToken(ctx, token2.Token)
		assert.Error(t, err2)
	})
}
