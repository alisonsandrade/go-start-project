package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/internal/auth/domain"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService implementa a interface auth.AuthService
type MockAuthService struct{ mock.Mock }

func (m *MockAuthService) Register(ctx context.Context, dto domain.RegisterRequest) (*domain.AuthResponseDTO, error) {
	args := m.Called(ctx, dto)
	if r := args.Get(0); r != nil {
		return r.(*domain.AuthResponseDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, dto domain.LoginRequest) (*domain.AuthResponseDTO, error) {
	args := m.Called(ctx, dto)
	if r := args.Get(0); r != nil {
		return r.(*domain.AuthResponseDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) RefreshSession(ctx context.Context, refreshToken string) (*domain.AuthResponseDTO, error) {
	args := m.Called(ctx, refreshToken)
	if r := args.Get(0); r != nil {
		return r.(*domain.AuthResponseDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *MockAuthService) ForgotPassword(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}

func (m *MockAuthService) ResetPassword(ctx context.Context, rawToken string, newPassword string) error {
	return m.Called(ctx, rawToken, newPassword).Error(0)
}

func (m *MockAuthService) ChangePassword(ctx context.Context, userID uuid.UUID, dto domain.ChangePasswordDTO) error {
	return m.Called(ctx, userID, dto).Error(0)
}

func TestAuthHandler_Register(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := auth.NewAuthHandler(mockSvc)

	t.Run("400 quando o body do json e malformado", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte("{invalid-json")))
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("400 quando a validacao da request falha (senha curta)", func(t *testing.T) {
		payload, _ := json.Marshal(domain.RegisterRequest{
			Name:     "Alison",
			Email:    "alison@example.com",
			Password: "123", // menos de 8 caracteres
		})

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("409 quando o email ja esta cadastrado", func(t *testing.T) {
		dto := domain.RegisterRequest{
			Name:     "Alison Silva",
			Email:    "duplicado@example.com",
			Password: "securePassword123!",
		}
		payload, _ := json.Marshal(dto)

		mockSvc.On("Register", mock.Anything, dto).Return(nil, auth.ErrEmailAlreadyExists).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("201 quando o cadastro e concluido com sucesso", func(t *testing.T) {
		dto := domain.RegisterRequest{
			Name:     "Alison Silva",
			Email:    "novo@example.com",
			Password: "securePassword123!",
		}
		payload, _ := json.Marshal(dto)

		mockSvc.On("Register", mock.Anything, dto).Return(&domain.AuthResponseDTO{
			AccessToken:  "access.jwt.token",
			RefreshToken: "refresh.token",
		}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := auth.NewAuthHandler(mockSvc)

	t.Run("400 quando o body e invalido", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("{invalid-json")))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("401 quando credenciais sao invalidas", func(t *testing.T) {
		dto := domain.LoginRequest{Email: "user@example.com", Password: "wrong"}
		payload, _ := json.Marshal(dto)

		mockSvc.On("Login", mock.Anything, dto).Return(nil, auth.ErrInvalidCredentials).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("200 quando login for bem sucedido", func(t *testing.T) {
		dto := domain.LoginRequest{Email: "user@example.com", Password: "correct"}
		payload, _ := json.Marshal(dto)

		mockSvc.On("Login", mock.Anything, dto).Return(&domain.AuthResponseDTO{
			AccessToken:  "token",
			RefreshToken: "token",
		}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := auth.NewAuthHandler(mockSvc)

	t.Run("400 quando o token nao for enviado", func(t *testing.T) {
		payload, _ := json.Marshal(domain.RefreshTokenDTO{RefreshToken: ""})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.RefreshToken(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("200 quando token for renovado", func(t *testing.T) {
		dto := domain.RefreshTokenDTO{RefreshToken: "valid-token"}
		payload, _ := json.Marshal(dto)

		mockSvc.On("RefreshSession", mock.Anything, "valid-token").Return(&domain.AuthResponseDTO{
			AccessToken: "novo-token",
		}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.RefreshToken(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestAuthHandler_ProtectedRoutes(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := auth.NewAuthHandler(mockSvc)

	t.Run("Logout retorna 401 se nao houver claims no contexto", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		rr := httptest.NewRecorder()

		handler.Logout(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Logout com sucesso retorna 200", func(t *testing.T) {
		userID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		ctx := context.WithValue(req.Context(), auth.UserClaimsKey, &token.CustomClaims{UserID: userID})
		req = req.WithContext(ctx)

		mockSvc.On("Logout", mock.Anything, userID).Return(nil).Once()

		rr := httptest.NewRecorder()
		handler.Logout(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("ChangePassword retorna 401 com senha atual incorreta", func(t *testing.T) {
		userID := uuid.New()
		dto := domain.ChangePasswordDTO{
			CurrentPassword: "wrong",
			NewPassword:     "newSecurePassword123!",
		}
		payload, _ := json.Marshal(dto)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader(payload))
		ctx := context.WithValue(req.Context(), auth.UserClaimsKey, &token.CustomClaims{UserID: userID})
		req = req.WithContext(ctx)

		mockSvc.On("ChangePassword", mock.Anything, userID, dto).Return(auth.ErrCurrentPasswordIncorrect).Once()

		rr := httptest.NewRecorder()
		handler.ChangePassword(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("ChangePassword retorna 401 sem claims", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader([]byte(`{}`)))
		rr := httptest.NewRecorder()

		handler.ChangePassword(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("ChangePassword retorna 200 com sucesso", func(t *testing.T) {
		userID := uuid.New()
		dto := domain.ChangePasswordDTO{CurrentPassword: "current", NewPassword: "newSecurePassword123!"}
		payload, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader(payload))
		ctx := context.WithValue(req.Context(), auth.UserClaimsKey, &token.CustomClaims{UserID: userID})
		req = req.WithContext(ctx)

		mockSvc.On("ChangePassword", mock.Anything, userID, dto).Return(nil).Once()

		rr := httptest.NewRecorder()
		handler.ChangePassword(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestAuthHandler_ForgotAndResetPassword(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := auth.NewAuthHandler(mockSvc)

	t.Run("ForgotPassword retorna 200 com mensagem padrao", func(t *testing.T) {
		payload, _ := json.Marshal(domain.ForgotPasswordDTO{Email: "user@example.com"})
		mockSvc.On("ForgotPassword", mock.Anything, "user@example.com").Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/forgot-password", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.ForgotPassword(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("ResetPassword retorna 400 com token invalido", func(t *testing.T) {
		dto := domain.ResetPasswordDTO{Token: "bad-token", NewPassword: "novaSenha123!"}
		payload, _ := json.Marshal(dto)

		mockSvc.On("ResetPassword", mock.Anything, dto.Token, dto.NewPassword).Return(auth.ErrResetTokenInvalid).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.ResetPassword(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("ResetPassword retorna 200 com sucesso", func(t *testing.T) {
		dto := domain.ResetPasswordDTO{Token: "good-token", NewPassword: "novaSenha123!"}
		payload, _ := json.Marshal(dto)

		mockSvc.On("ResetPassword", mock.Anything, dto.Token, dto.NewPassword).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.ResetPassword(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestAuthHandler_AuthRoutes(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := auth.NewAuthHandler(mockSvc)

	router := handler.AuthRoutes(testConfig)
	assert.NotNil(t, router)
}

func TestAuthHandler_ErrorBranches(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := auth.NewAuthHandler(mockSvc)

	t.Run("Register retorna 500 em erro generico do servico", func(t *testing.T) {
		dto := domain.RegisterRequest{
			Name:     "Alison Silva",
			Email:    "erro@example.com",
			Password: "securePassword123!",
		}
		payload, _ := json.Marshal(dto)
		mockSvc.On("Register", mock.Anything, dto).Return(nil, errors.New("db error")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.Register(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("Login retorna 500 em erro generico", func(t *testing.T) {
		dto := domain.LoginRequest{Email: "user@example.com", Password: "pwd"}
		payload, _ := json.Marshal(dto)
		mockSvc.On("Login", mock.Anything, dto).Return(nil, errors.New("generic error")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("RefreshToken retorna 401 em erro do servico", func(t *testing.T) {
		dto := domain.RefreshTokenDTO{RefreshToken: "expired"}
		payload, _ := json.Marshal(dto)
		mockSvc.On("RefreshSession", mock.Anything, "expired").Return(nil, auth.ErrInvalidRefreshToken).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.RefreshToken(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("RefreshToken return error 400 when invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
		rr := httptest.NewRecorder()

		handler.RefreshToken(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Logout retorna 500 em falha do servico", func(t *testing.T) {
		userID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		ctx := context.WithValue(req.Context(), auth.UserClaimsKey, &token.CustomClaims{UserID: userID})
		req = req.WithContext(ctx)

		mockSvc.On("Logout", mock.Anything, userID).Return(errors.New("db down")).Once()

		rr := httptest.NewRecorder()
		handler.Logout(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("ForgotPassword retorna 500 em falha de servico", func(t *testing.T) {
		payload, _ := json.Marshal(domain.ForgotPasswordDTO{Email: "user@example.com"})
		mockSvc.On("ForgotPassword", mock.Anything, "user@example.com").Return(errors.New("mailer down")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/forgot-password", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.ForgotPassword(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("ForgotPassword return error 400 when invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/forgot-password", nil)
		rr := httptest.NewRecorder()

		handler.ForgotPassword(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("ResetPassword retorna 400 quando token expirou", func(t *testing.T) {
		dto := domain.ResetPasswordDTO{Token: "exp-token", NewPassword: "novaSenha123!"}
		payload, _ := json.Marshal(dto)
		mockSvc.On("ResetPassword", mock.Anything, dto.Token, dto.NewPassword).Return(auth.ErrResetTokenExpired).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.ResetPassword(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("ResetPassword returns error 400 when invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", nil)
		rr := httptest.NewRecorder()

		handler.ResetPassword(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("ResetPassord returns error 400 when it doesn't meet the Value Object", func(t *testing.T) {
		dto := domain.ResetPasswordDTO{Token: "token", NewPassword: "short"}
		payload, _ := json.Marshal(dto)

		mockSvc.On("ResetPassword", mock.Anything, dto.Token, dto.NewPassword).
			Return(errors.New("password must be at least 8 characteres long")).
			Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.ResetPassword(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("ChangePassword retorna 400 quando falhar por validacao de senha", func(t *testing.T) {
		userID := uuid.New()
		dto := domain.ChangePasswordDTO{CurrentPassword: "current", NewPassword: "curta"}
		payload, _ := json.Marshal(dto)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader(payload))
		ctx := context.WithValue(req.Context(), auth.UserClaimsKey, &token.CustomClaims{UserID: userID})
		req = req.WithContext(ctx)

		mockSvc.On("ChangePassword", mock.Anything, userID, dto).Return(errors.New("password too short")).Once()

		rr := httptest.NewRecorder()
		handler.ChangePassword(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("ChangePassword return error 400 when invalid request body", func(t *testing.T) {
		userID := uuid.New()
		bodyInvalid := []byte(`{json-bad-formated`)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader(bodyInvalid))
		ctx := context.WithValue(req.Context(), auth.UserClaimsKey, &token.CustomClaims{UserID: userID})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.ChangePassword(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockSvc.AssertExpectations(t)
	})
}
