// Package auth service
package auth

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/auth/domain"
	"github.com/alisonsandrade/go-start-project/internal/config"
	usersDomain "github.com/alisonsandrade/go-start-project/internal/users/domain"
	pkgDomain "github.com/alisonsandrade/go-start-project/pkg/domain"
	"github.com/alisonsandrade/go-start-project/pkg/mailer"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, dto domain.RegisterRequest) (*domain.AuthResponseDTO, error)
	Login(ctx context.Context, dto domain.LoginRequest) (*domain.AuthResponseDTO, error)
	RefreshSession(ctx context.Context, refreshToken string) (*domain.AuthResponseDTO, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token string, newPassword string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, dto domain.ChangePasswordDTO) error
}

type authService struct {
	userRepo  UserRepository
	tokenRepo TokenRepository
	cfg       *config.Config
	mailer    mailer.Mailer
}

func NewAuthService(
	userRepo UserRepository,
	tokenRepo TokenRepository,
	cfg *config.Config,
	mailer mailer.Mailer,
) AuthService {
	return &authService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
		mailer:    mailer,
	}
}

// Register creates a new user. Public registration always assigns RoleUser.
func (s *authService) Register(ctx context.Context, dto domain.RegisterRequest) (*domain.AuthResponseDTO, error) {
	existingUser, err := s.userRepo.FindByEmail(ctx, dto.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	email, err := pkgDomain.NewEmail(dto.Email)
	if err != nil {
		return nil, err
	}

	password, err := pkgDomain.NewPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	userRoleDefaultID, err := s.userRepo.GetDefaultRoleID(ctx)
	if err != nil {
		return nil, err
	}

	user := &usersDomain.User{
		Name:      dto.Name,
		Email:     email,
		Password:  password,
		RoleID:    userRoleDefaultID,
		Phone:     dto.Phone,
		AvatarURL: dto.AvatarURL,
		JobTitle:  dto.JobTitle,
		Bio:       dto.Bio,
		IsActive:  true,
	}

	// if err := user.HashPassword(); err != nil {
	// 	return nil, err
	// }

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(ctx, user)
}

// Login authenticates the given credentials and returns a new token pair.
func (s *authService) Login(ctx context.Context, dto domain.LoginRequest) (*domain.AuthResponseDTO, error) {
	email, err := pkgDomain.NewEmail(dto.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByEmail(ctx, email.String())
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if err := user.Password.Compare(dto.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateAuthResponse(ctx, user)
}

// Logout invalidates the user's session by deleting all of their refresh tokens.
func (s *authService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.tokenRepo.DeleteByUserID(ctx, userID)
}

// RefreshSession validates the refresh token, rotates it, and returns a new token pair.
func (s *authService) RefreshSession(ctx context.Context, refreshToken string) (*domain.AuthResponseDTO, error) {
	rt, err := s.tokenRepo.FindByToken(ctx, refreshToken)
	if err != nil || rt == nil {
		return nil, ErrInvalidRefreshToken
	}

	if time.Now().After(rt.ExpiresAt) {
		_ = s.tokenRepo.Delete(ctx, refreshToken)
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.FindByID(ctx, rt.UserID)
	if err != nil || user == nil {
		return nil, ErrInvalidRefreshToken
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Rotation invalid old refresh token
	if err := s.tokenRepo.Delete(ctx, refreshToken); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(ctx, user)
}

// generateAuthResponse builds the access and refresh tokens and persists the refresh token.
func (s *authService) generateAuthResponse(ctx context.Context, user *usersDomain.User) (*domain.AuthResponseDTO, error) {
	email, err := pkgDomain.NewEmail(user.Email.String())
	if err != nil {
		return nil, err
	}

	expHours, _ := strconv.Atoi(s.cfg.JWTExpirationHours)
	if expHours == 0 {
		expHours = 24
	}

	accessToken, err := token.GenerateToken(
		user.ID, email.String(), user.RoleID, s.cfg.JWTSecret, expHours,
	)
	if err != nil {
		return nil, err
	}

	refreshTokenStr, err := token.GenerateSecureToken()
	if err != nil {
		return nil, err
	}

	rt := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7), // 7 dias
	}
	if err := s.tokenRepo.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &domain.AuthResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		User:         *user,
	}, nil
}

// ForgotPassword return a code for user autenticated again on system
func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return nil
	}

	rawToken, err := token.GenerateSecureToken()
	if err != nil {
		return err
	}

	// Cria a entidade do seu domínio usando a struct que você já tem
	resetToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     rawToken,
		ExpiresAt: time.Now().UTC().Add(15 * time.Minute),
	}

	if err := s.tokenRepo.Create(ctx, resetToken); err != nil {
		return err
	}

	if err := s.mailer.SendPasswordReset(context.Background(), user.Email.String(), resetToken.Token); err != nil {
		log.Printf("Falha ao enfileirar e-mail: %v", err)
	}

	return nil
}

// ResetPassword validates the token, updates the password and revokes all active sessions.
func (s *authService) ResetPassword(ctx context.Context, rawToken string, rawNewPassword string) error {
	resetToken, err := s.tokenRepo.FindByToken(ctx, rawToken)
	if err != nil || resetToken == nil {
		return ErrResetTokenInvalid
	}

	if time.Now().UTC().After(resetToken.ExpiresAt) {
		return ErrResetTokenExpired
	}

	newPassword, err := pkgDomain.NewPassword(rawNewPassword)
	if err != nil {
		return err
	}

	user, err := s.userRepo.FindByID(ctx, resetToken.UserID)
	if err != nil || user == nil {
		return ErrResetTokenInvalid
	}

	user.Password = newPassword

	// save new password the user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Revokes all tokens that user logged
	if err := s.tokenRepo.DeleteByUserID(ctx, resetToken.UserID); err != nil {
		return err
	}

	return nil
}

func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, dto domain.ChangePasswordDTO) error {
	// 1. Busca o usuário pelo ID vindo das claims do token JWT
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return ErrInvalidCredentials
	}

	// 2. Valida se a senha atual confere usando a comparação do Value Object
	if err := user.Password.Compare(dto.CurrentPassword); err != nil {
		return ErrCurrentPasswordIncorrect
	}

	// 3. Cria o Value Object da nova senha (valida regras de tamanho e gera hash)
	newPassword, err := pkgDomain.NewPassword(dto.NewPassword)
	if err != nil {
		return err
	}

	// 4. Atualiza o usuário
	user.Password = newPassword
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// 5. Boa prática de segurança: revoga os refresh tokens existentes para deslogar outras sessões
	_ = s.tokenRepo.DeleteByUserID(ctx, userID)

	return nil
}
