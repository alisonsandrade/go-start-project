// Package auth provide the auth's domains and services.
package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/internal/auth/domain"
	"github.com/alisonsandrade/go-start-project/internal/config"
	usersDomain "github.com/alisonsandrade/go-start-project/internal/users/domain"
	pkgDomain "github.com/alisonsandrade/go-start-project/pkg/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) Create(ctx context.Context, u *usersDomain.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *MockUserRepo) Update(ctx context.Context, u *usersDomain.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*usersDomain.User, error) {
	args := m.Called(ctx, email)
	if u := args.Get(0); u != nil {
		return u.(*usersDomain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*usersDomain.User, error) {
	args := m.Called(ctx, id)
	if u := args.Get(0); u != nil {
		return u.(*usersDomain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepo) GetDefaultRoleID(ctx context.Context) (uuid.UUID, error) {
	args := m.Called(ctx)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

type MockTokenRepo struct{ mock.Mock }

func (m *MockTokenRepo) Create(ctx context.Context, t *domain.RefreshToken) error {
	return m.Called(ctx, t).Error(0)
}

func (m *MockTokenRepo) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, token)
	if t := args.Get(0); t != nil {
		return t.(*domain.RefreshToken), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTokenRepo) DeleteByUserID(ctx context.Context, uid uuid.UUID) error {
	return m.Called(ctx, uid).Error(0)
}

func (m *MockTokenRepo) Delete(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}

type MockMailer struct{ mock.Mock }

func (m *MockMailer) SendPasswordReset(ctx context.Context, toEmail, resetToken string) error {
	return m.Called(ctx, toEmail, resetToken).Error(0)
}
func (m *MockMailer) Close() error { return nil }

var testConfig = &config.Config{
	JWTSecret:          "minha-chave-secreta-de-testes-32-bits!",
	JWTExpirationHours: "24",
}

// --- Testes ---

func TestAuthService_Register(t *testing.T) {
	t.Run("sucesso ao registrar", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		mailer := new(MockMailer)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, mailer)

		req := domain.RegisterRequest{
			Name:     "Alison Silva",
			Email:    "alison@example.com",
			Password: "strongPassword123!",
		}
		roleID := uuid.New()

		userRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, nil)
		userRepo.On("GetDefaultRoleID", mock.Anything).Return(roleID, nil)
		userRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
		tokenRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

		resp, err := svc.Register(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
	})

	t.Run("falha quando email ja existe", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)

		req := domain.RegisterRequest{Email: "existente@example.com"}
		userRepo.On("FindByEmail", mock.Anything, req.Email).Return(&usersDomain.User{}, nil)

		resp, err := svc.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, auth.ErrEmailAlreadyExists)
	})

	t.Run("fail when repository returns an error get email", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)

		req := domain.RegisterRequest{Email: "test@example.com"}

		userRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, errors.New("database down"))
		resp, err := svc.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "database down")
	})

	t.Run("falha quando o email e invalido pelo value object", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)
		req := domain.RegisterRequest{Email: "email-invalido"}

		userRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, nil)

		resp, err := svc.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.Error(t, err)
	})

	t.Run("falha quando a senha e invalida pelo value object", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)

		req := domain.RegisterRequest{
			Email:    "valido@example.com",
			Password: "curta",
		}

		userRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, nil)

		resp, err := svc.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.Error(t, err)
	})

	t.Run("falha quando nao consegue buscar a role padrao", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)

		req := domain.RegisterRequest{
			Email:    "valido@example.com",
			Password: "SenhaMuitoForte123!",
		}

		userRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, nil)
		userRepo.On("GetDefaultRoleID", mock.Anything).Return(uuid.Nil, errors.New("role not found"))

		resp, err := svc.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "role not found")
	})

	t.Run("fail trying create a new user", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)

		req := domain.RegisterRequest{
			Name:     "Alison Silva",
			Email:    "alison@example.com",
			Password: "strongPassword123!",
		}

		userRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, nil)
		userRepo.On("GetDefaultRoleID", mock.Anything).Return(uuid.Nil, nil)
		userRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(errors.New("error create a new user"))

		resp, err := svc.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "error create a new user")
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Run("falha com email invalido", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)

		resp, err := svc.Login(context.Background(), domain.LoginRequest{
			Email: "email-invalido", Password: "qualquercoisa",
		})

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	})

	t.Run("sucesso ao logar com credenciais validas", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, nil)

		email, _ := pkgDomain.NewEmail("valid@example.com")
		pwd, _ := pkgDomain.NewPassword("correctPassword123!")
		user := &usersDomain.User{
			ID:       uuid.New(),
			Email:    email,
			Password: pwd,
			IsActive: true,
		}

		userRepo.On("FindByEmail", mock.Anything, email.String()).Return(user, nil)
		tokenRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

		resp, err := svc.Login(context.Background(), domain.LoginRequest{
			Email:    email.String(),
			Password: "correctPassword123!",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.AccessToken)
	})

	t.Run("falha com credenciais invalidas", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)

		userRepo.On("FindByEmail", mock.Anything, "inexistente@example.com").Return(nil, nil)

		resp, err := svc.Login(context.Background(), domain.LoginRequest{
			Email:    "inexistente@example.com",
			Password: "qualquercoisa",
		})

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	})

	t.Run("falha com usuario inativo", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)

		email, _ := pkgDomain.NewEmail("inativo@example.com")
		pwd, _ := pkgDomain.NewPassword("pass12345!")
		user := &usersDomain.User{Email: email, Password: pwd, IsActive: false}

		userRepo.On("FindByEmail", mock.Anything, email.String()).Return(user, nil)

		resp, err := svc.Login(context.Background(), domain.LoginRequest{
			Email:    email.String(),
			Password: "pass12345!",
		})

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, auth.ErrUserInactive)
	})

	t.Run("falha com senha errada", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)

		email, _ := pkgDomain.NewEmail("user@example.com")
		pwd, _ := pkgDomain.NewPassword("correta123!")
		user := &usersDomain.User{Email: email, Password: pwd, IsActive: true}

		userRepo.On("FindByEmail", mock.Anything, email.String()).Return(user, nil)

		resp, err := svc.Login(context.Background(), domain.LoginRequest{
			Email:    email.String(),
			Password: "senha-completamente-errada",
		})

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	})

	t.Run("falha quando nao consegue persistir o refresh token", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, nil)
		email, _ := pkgDomain.NewEmail("persist@example.com")
		pwd, _ := pkgDomain.NewPassword("correctPassword123!")
		user := &usersDomain.User{ID: uuid.New(), Email: email, Password: pwd, IsActive: true}
		persistErr := errors.New("refresh token persistence failed")

		userRepo.On("FindByEmail", mock.Anything, email.String()).Return(user, nil)
		tokenRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(persistErr)

		resp, err := svc.Login(context.Background(), domain.LoginRequest{
			Email: email.String(), Password: "correctPassword123!",
		})

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, persistErr)
	})
}

func TestAuthService_Logout(t *testing.T) {
	tokenRepo := new(MockTokenRepo)
	svc := auth.NewAuthService(nil, tokenRepo, testConfig, nil)
	userID := uuid.New()

	tokenRepo.On("DeleteByUserID", mock.Anything, userID).Return(nil)

	err := svc.Logout(context.Background(), userID)
	assert.NoError(t, err)
	tokenRepo.AssertExpectations(t)
}

func TestAuthService_RefreshSession(t *testing.T) {
	t.Run("falha quando o token nao existe", func(t *testing.T) {
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(nil, tokenRepo, testConfig, nil)
		tokenRepo.On("FindByToken", mock.Anything, "missing-token").Return(nil, nil)

		resp, err := svc.RefreshSession(context.Background(), "missing-token")

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, auth.ErrInvalidRefreshToken)
	})

	t.Run("sucesso ao rotacionar refresh token", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, nil)

		userID := uuid.New()
		rawToken := "valid-refresh-token"
		existingToken := &domain.RefreshToken{
			UserID:    userID,
			Token:     rawToken,
			ExpiresAt: time.Now().Add(2 * time.Hour),
		}

		email, _ := pkgDomain.NewEmail("user@example.com")
		pwd, _ := pkgDomain.NewPassword("password123!")
		user := &usersDomain.User{ID: userID, Email: email, Password: pwd, IsActive: true}

		tokenRepo.On("FindByToken", mock.Anything, rawToken).Return(existingToken, nil)
		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
		tokenRepo.On("Delete", mock.Anything, rawToken).Return(nil)
		tokenRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

		resp, err := svc.RefreshSession(context.Background(), rawToken)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("falha com token expirado", func(t *testing.T) {
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(nil, tokenRepo, testConfig, nil)

		rawToken := "expired-token"
		tokenRepo.On("FindByToken", mock.Anything, rawToken).Return(&domain.RefreshToken{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}, nil)
		tokenRepo.On("Delete", mock.Anything, rawToken).Return(nil)

		resp, err := svc.RefreshSession(context.Background(), rawToken)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, auth.ErrInvalidRefreshToken)
	})

	t.Run("falha quando usuario esta inativo", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, nil)
		userID := uuid.New()
		rawToken := "inactive-user-token"
		rt := &domain.RefreshToken{UserID: userID, Token: rawToken, ExpiresAt: time.Now().Add(time.Hour)}
		email, _ := pkgDomain.NewEmail("inactive@example.com")
		user := &usersDomain.User{ID: userID, Email: email, IsActive: false}

		tokenRepo.On("FindByToken", mock.Anything, rawToken).Return(rt, nil)
		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)

		resp, err := svc.RefreshSession(context.Background(), rawToken)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, auth.ErrUserInactive)
	})

	t.Run("falha quando a rotacao nao consegue apagar o token antigo", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, nil)
		userID := uuid.New()
		rawToken := "rotation-error-token"
		rt := &domain.RefreshToken{UserID: userID, Token: rawToken, ExpiresAt: time.Now().Add(time.Hour)}
		email, _ := pkgDomain.NewEmail("rotation@example.com")
		pwd, _ := pkgDomain.NewPassword("password123!")
		user := &usersDomain.User{ID: userID, Email: email, Password: pwd, IsActive: true}
		deleteErr := errors.New("delete old token failed")

		tokenRepo.On("FindByToken", mock.Anything, rawToken).Return(rt, nil)
		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
		tokenRepo.On("Delete", mock.Anything, rawToken).Return(deleteErr)

		resp, err := svc.RefreshSession(context.Background(), rawToken)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, deleteErr)
	})

	t.Run("falha quando o email persistido do usuario e invalido", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, nil)
		userID := uuid.New()
		rawToken := "invalid-email-token"
		rt := &domain.RefreshToken{UserID: userID, Token: rawToken, ExpiresAt: time.Now().Add(time.Hour)}
		user := &usersDomain.User{ID: userID, IsActive: true}

		tokenRepo.On("FindByToken", mock.Anything, rawToken).Return(rt, nil)
		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
		tokenRepo.On("Delete", mock.Anything, rawToken).Return(nil)

		resp, err := svc.RefreshSession(context.Background(), rawToken)

		assert.Nil(t, resp)
		assert.Error(t, err)
	})
}

func TestAuthService_ForgotPassword(t *testing.T) {
	t.Run("falha quando nao consegue criar o token de recuperacao", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		mailer := new(MockMailer)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, mailer)
		email, _ := pkgDomain.NewEmail("create-error@example.com")
		user := &usersDomain.User{ID: uuid.New(), Email: email}
		createErr := errors.New("cannot create reset token")

		userRepo.On("FindByEmail", mock.Anything, email.String()).Return(user, nil)
		tokenRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(createErr)

		err := svc.ForgotPassword(context.Background(), email.String())

		assert.ErrorIs(t, err, createErr)
	})

	t.Run("sucesso ao solicitar recuperação", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		mailer := new(MockMailer)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, mailer)

		emailStr := "user@example.com"
		email, _ := pkgDomain.NewEmail(emailStr)
		user := &usersDomain.User{ID: uuid.New(), Email: email}

		userRepo.On("FindByEmail", mock.Anything, emailStr).Return(user, nil)
		tokenRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)
		mailer.On("SendPasswordReset", mock.Anything, emailStr, mock.AnythingOfType("string")).Return(nil)

		err := svc.ForgotPassword(context.Background(), emailStr)
		assert.NoError(t, err)
		mailer.AssertExpectations(t)
	})

	t.Run("não retorna erro caso o e-mail não seja encontrado", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)

		userRepo.On("FindByEmail", mock.Anything, "naoexiste@example.com").Return(nil, errors.New("not found"))

		err := svc.ForgotPassword(context.Background(), "naoexiste@example.com")
		assert.NoError(t, err)
	})

	t.Run("nao falha quando o envio do email retorna erro", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		mailer := new(MockMailer)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, mailer)
		email, _ := pkgDomain.NewEmail("mailer-error@example.com")
		user := &usersDomain.User{ID: uuid.New(), Email: email}

		userRepo.On("FindByEmail", mock.Anything, email.String()).Return(user, nil)
		tokenRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)
		mailer.On("SendPasswordReset", mock.Anything, email.String(), mock.AnythingOfType("string")).Return(errors.New("mailer unavailable"))

		err := svc.ForgotPassword(context.Background(), email.String())

		assert.NoError(t, err)
	})
}

func TestAuthService_ResetPassword(t *testing.T) {
	t.Run("sucesso ao redefinir senha", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, nil)

		userID := uuid.New()
		rawToken := "reset-token"
		resetToken := &domain.RefreshToken{
			UserID:    userID,
			Token:     rawToken,
			ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		}

		email, _ := pkgDomain.NewEmail("user@example.com")
		user := &usersDomain.User{ID: userID, Email: email}

		tokenRepo.On("FindByToken", mock.Anything, rawToken).Return(resetToken, nil)
		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
		userRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
		tokenRepo.On("DeleteByUserID", mock.Anything, userID).Return(nil)

		err := svc.ResetPassword(context.Background(), rawToken, "NovaSenha123!")
		assert.NoError(t, err)
	})

	t.Run("falha com token inválido", func(t *testing.T) {
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(nil, tokenRepo, testConfig, nil)

		tokenRepo.On("FindByToken", mock.Anything, "invalido").Return(nil, errors.New("not found"))

		err := svc.ResetPassword(context.Background(), "invalido", "NovaSenha123!")
		assert.ErrorIs(t, err, auth.ErrResetTokenInvalid)
	})

	t.Run("falha com token expirado", func(t *testing.T) {
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(nil, tokenRepo, testConfig, nil)

		tokenRepo.On("FindByToken", mock.Anything, "expirado").Return(&domain.RefreshToken{
			ExpiresAt: time.Now().UTC().Add(-5 * time.Minute),
		}, nil)

		err := svc.ResetPassword(context.Background(), "expirado", "NovaSenha123!")
		assert.ErrorIs(t, err, auth.ErrResetTokenExpired)
	})

	t.Run("falha com nova senha invalida", func(t *testing.T) {
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(nil, tokenRepo, testConfig, nil)
		rawToken := "invalid-password-token"
		tokenRepo.On("FindByToken", mock.Anything, rawToken).Return(&domain.RefreshToken{
			UserID: uuid.New(), ExpiresAt: time.Now().UTC().Add(time.Minute),
		}, nil)

		err := svc.ResetPassword(context.Background(), rawToken, "curta")

		assert.Error(t, err)
	})

	t.Run("falha ao atualizar a senha", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, nil)
		userID := uuid.New()
		rawToken := "update-error-token"
		email, _ := pkgDomain.NewEmail("update-error@example.com")
		user := &usersDomain.User{ID: userID, Email: email}
		updateErr := errors.New("cannot update password")

		tokenRepo.On("FindByToken", mock.Anything, rawToken).Return(&domain.RefreshToken{
			UserID: userID, ExpiresAt: time.Now().UTC().Add(time.Minute),
		}, nil)
		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
		userRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(updateErr)

		err := svc.ResetPassword(context.Background(), rawToken, "NovaSenha123!")

		assert.ErrorIs(t, err, updateErr)
	})

	t.Run("falha ao revogar sessoes depois de atualizar a senha", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, nil)
		userID := uuid.New()
		rawToken := "revoke-error-token"
		email, _ := pkgDomain.NewEmail("revoke-error@example.com")
		user := &usersDomain.User{ID: userID, Email: email}
		revokeErr := errors.New("cannot revoke sessions")

		tokenRepo.On("FindByToken", mock.Anything, rawToken).Return(&domain.RefreshToken{
			UserID: userID, ExpiresAt: time.Now().UTC().Add(time.Minute),
		}, nil)
		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
		userRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
		tokenRepo.On("DeleteByUserID", mock.Anything, userID).Return(revokeErr)

		err := svc.ResetPassword(context.Background(), rawToken, "NovaSenha123!")

		assert.ErrorIs(t, err, revokeErr)
	})
}

func TestAuthService_ChangePassword(t *testing.T) {
	t.Run("falha quando usuario nao e encontrado", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)
		userID := uuid.New()
		userRepo.On("FindByID", mock.Anything, userID).Return(nil, errors.New("user not found"))

		err := svc.ChangePassword(context.Background(), userID, domain.ChangePasswordDTO{})

		assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	})

	t.Run("sucesso ao trocar senha autenticado", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		tokenRepo := new(MockTokenRepo)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, nil)

		userID := uuid.New()
		pwd, _ := pkgDomain.NewPassword("senhaAtual123!")
		user := &usersDomain.User{ID: userID, Password: pwd}

		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
		userRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
		tokenRepo.On("DeleteByUserID", mock.Anything, userID).Return(nil)

		err := svc.ChangePassword(context.Background(), userID, domain.ChangePasswordDTO{
			CurrentPassword: "senhaAtual123!",
			NewPassword:     "novaSenhaSegura123!",
		})

		assert.NoError(t, err)
	})

	t.Run("falha quando a senha atual está incorreta", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)

		userID := uuid.New()
		pwd, _ := pkgDomain.NewPassword("senhaAtual123!")
		user := &usersDomain.User{ID: userID, Password: pwd}

		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)

		err := svc.ChangePassword(context.Background(), userID, domain.ChangePasswordDTO{
			CurrentPassword: "senhaErrada",
			NewPassword:     "novaSenhaSegura123!",
		})

		assert.ErrorIs(t, err, auth.ErrCurrentPasswordIncorrect)
	})

	t.Run("falha quando a nova senha e invalida", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)
		userID := uuid.New()
		pwd, _ := pkgDomain.NewPassword("senhaAtual123!")
		user := &usersDomain.User{ID: userID, Password: pwd}
		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)

		err := svc.ChangePassword(context.Background(), userID, domain.ChangePasswordDTO{
			CurrentPassword: "senhaAtual123!", NewPassword: "curta",
		})

		assert.Error(t, err)
	})

	t.Run("falha ao persistir a nova senha", func(t *testing.T) {
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, nil, testConfig, nil)
		userID := uuid.New()
		pwd, _ := pkgDomain.NewPassword("senhaAtual123!")
		user := &usersDomain.User{ID: userID, Password: pwd}
		updateErr := errors.New("cannot update user")
		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
		userRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(updateErr)

		err := svc.ChangePassword(context.Background(), userID, domain.ChangePasswordDTO{
			CurrentPassword: "senhaAtual123!", NewPassword: "novaSenhaSegura123!",
		})

		assert.ErrorIs(t, err, updateErr)
	})
}

func TestAuthService_EdgeCases(t *testing.T) {
	t.Run("RefreshSession falha se usuario nao for encontrado ou estiver inativo", func(t *testing.T) {
		tokenRepo := new(MockTokenRepo)
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, nil)

		userID := uuid.New()
		rt := &domain.RefreshToken{
			UserID:    userID,
			Token:     "token-abc",
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		tokenRepo.On("FindByToken", mock.Anything, "token-abc").Return(rt, nil)
		userRepo.On("FindByID", mock.Anything, userID).Return(nil, errors.New("user not found"))

		resp, err := svc.RefreshSession(context.Background(), "token-abc")
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, auth.ErrInvalidRefreshToken)
	})

	t.Run("ResetPassword falha se usuario associado ao token nao for encontrado", func(t *testing.T) {
		tokenRepo := new(MockTokenRepo)
		userRepo := new(MockUserRepo)
		svc := auth.NewAuthService(userRepo, tokenRepo, testConfig, nil)

		userID := uuid.New()
		rt := &domain.RefreshToken{
			UserID:    userID,
			Token:     "token-xyz",
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}

		tokenRepo.On("FindByToken", mock.Anything, "token-xyz").Return(rt, nil)
		userRepo.On("FindByID", mock.Anything, userID).Return(nil, errors.New("not found"))

		err := svc.ResetPassword(context.Background(), "token-xyz", "StrongPassword123!")
		assert.ErrorIs(t, err, auth.ErrResetTokenInvalid)
	})
}
