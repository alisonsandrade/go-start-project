package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/internal/auth/domain"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupTokenTestDB inicializa um SQLite em memória compatível
func setupTokenTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE refresh_tokens (
			id TEXT PRIMARY KEY,
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
	ctx := context.Background()

	userID1 := uuid.New()
	userID2 := uuid.New()

	token1 := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID1,
		Token:     "token-abc-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	token2 := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID1,
		Token:     "token-def-456",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	token3 := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID2,
		Token:     "token-ghi-789",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	t.Run("Create e FindByToken com sucesso", func(t *testing.T) {
		err := repo.Create(ctx, token1)
		assert.NoError(t, err)

		found, err := repo.FindByToken(ctx, token1.Token)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, token1.ID, found.ID)
		assert.Equal(t, token1.UserID, found.UserID)
		assert.Equal(t, token1.Token, found.Token)
	})

	t.Run("FindByToken deve retornar erro quando token não existe", func(t *testing.T) {
		found, err := repo.FindByToken(ctx, "token-inexistente")
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("Delete deve remover um token específico", func(t *testing.T) {
		err := repo.Create(ctx, token3)
		assert.NoError(t, err)

		// Deleta token3
		err = repo.Delete(ctx, token3.Token)
		assert.NoError(t, err)

		// Confirma que não existe mais
		found, err := repo.FindByToken(ctx, token3.Token)
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("DeleteByUserID deve revogar todos os tokens de um usuário específico", func(t *testing.T) {
		// Insere o segundo token para o userID1
		err := repo.Create(ctx, token2)
		assert.NoError(t, err)

		// Exclui todos os tokens associados ao userID1 (token1 e token2)
		err = repo.DeleteByUserID(ctx, userID1)
		assert.NoError(t, err)

		// Confirma que nenhum dos dois tokens pode ser encontrado
		_, err1 := repo.FindByToken(ctx, token1.Token)
		assert.Error(t, err1)

		_, err2 := repo.FindByToken(ctx, token2.Token)
		assert.Error(t, err2)
	})
}
