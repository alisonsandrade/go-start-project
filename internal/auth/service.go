// Package auth service
package auth

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/auth/domain"
	"github.com/alisonsandrade/go-start-project/internal/config"
	baseDomain "github.com/alisonsandrade/go-start-project/internal/domain"
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
	tm        TransactionManager
	cfg       *config.Config
	mailer    mailer.Mailer
}

func NewAuthService(
	userRepo UserRepository,
	tokenRepo TokenRepository,
	cfg *config.Config,
	mailer mailer.Mailer,
	tm ...TransactionManager,
) AuthService {
	var transactionManager TransactionManager
	if len(tm) > 0 {
		transactionManager = tm[0]
	}
	return &authService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		tm:        transactionManager,
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

// RefreshSession validates the refresh token, rotates it, and returns a new token pair inside an atomic transaction.
func (s *authService) RefreshSession(ctx context.Context, refreshToken string) (*domain.AuthResponseDTO, error) {
	rt, err := s.tokenRepo.FindByToken(ctx, refreshToken)
	if err != nil || rt == nil {
		return nil, ErrInvalidRefreshToken
	}

	if time.Now().After(rt.ExpiresAt) {
		_ = s.tokenRepo.Delete(ctx, refreshToken)
		return nil, ErrInvalidRefreshToken
	}

	userCtx := ctx
	if ctx.Value(token.ClaimsContextKey) == nil && rt.TenantID != uuid.Nil {
		userCtx = context.WithValue(ctx, token.ClaimsContextKey, &token.CustomClaims{
			UserID:   rt.UserID,
			TenantID: rt.TenantID,
		})
	}

	user, err := s.userRepo.FindByID(userCtx, rt.UserID)
	if err != nil || user == nil {
		return nil, ErrInvalidRefreshToken
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	email, err := pkgDomain.NewEmail(user.Email.String())
	if err != nil {
		return nil, err
	}

	expHours, _ := strconv.Atoi(s.cfg.JWTExpirationHours)
	if expHours == 0 {
		expHours = 24
	}

	accessToken, err := token.GenerateToken(
		user.ID, email.String(), user.RoleID, s.cfg.JWTSecret, expHours, user.TenantID,
	)
	if err != nil {
		return nil, err
	}

	newRefreshTokenStr, err := token.GenerateSecureToken()
	if err != nil {
		return nil, err
	}

	newRT := &domain.RefreshToken{
		BaseModelTenant: baseDomain.BaseModelTenant{
			BaseModel: baseDomain.BaseModel{ID: uuid.New()},
			TenantID:  user.TenantID,
		},
		UserID:    user.ID,
		Token:     newRefreshTokenStr,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	rotateFn := func(txCtx context.Context) error {
		if err := s.tokenRepo.Delete(txCtx, refreshToken); err != nil {
			return err
		}
		return s.tokenRepo.Create(txCtx, newRT)
	}

	if s.tm != nil {
		if err := s.tm.Transaction(userCtx, rotateFn); err != nil {
			return nil, err
		}
	} else {
		if err := rotateFn(userCtx); err != nil {
			return nil, err
		}
	}

	return &domain.AuthResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenStr,
		User:         *user,
	}, nil
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
		user.ID, email.String(), user.RoleID, s.cfg.JWTSecret, expHours, user.TenantID,
	)
	if err != nil {
		return nil, err
	}

	refreshTokenStr, err := token.GenerateSecureToken()
	if err != nil {
		return nil, err
	}

	rt := &domain.RefreshToken{
		BaseModelTenant: baseDomain.BaseModelTenant{
			BaseModel: baseDomain.BaseModel{ID: uuid.New()},
			TenantID:  user.TenantID,
		},
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	tokenCtx := ctx
	if ctx.Value(token.ClaimsContextKey) == nil && user.TenantID != uuid.Nil {
		tokenCtx = context.WithValue(ctx, token.ClaimsContextKey, &token.CustomClaims{
			UserID:   user.ID,
			TenantID: user.TenantID,
		})
	}

	if err := s.tokenRepo.Create(tokenCtx, rt); err != nil {
		return nil, err
	}

	return &domain.AuthResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		User:         *user,
	}, nil
}

// ForgotPassword return a code for user authenticated again on system
func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return nil
	}

	rawToken, err := token.GenerateSecureToken()
	if err != nil {
		return err
	}

	resetToken := &domain.RefreshToken{
		BaseModelTenant: baseDomain.BaseModelTenant{
			BaseModel: baseDomain.BaseModel{ID: uuid.New()},
			TenantID:  user.TenantID,
		},
		UserID:    user.ID,
		Token:     rawToken,
		ExpiresAt: time.Now().UTC().Add(15 * time.Minute),
	}

	tokenCtx := ctx
	if ctx.Value(token.ClaimsContextKey) == nil && user.TenantID != uuid.Nil {
		tokenCtx = context.WithValue(ctx, token.ClaimsContextKey, &token.CustomClaims{
			UserID:   user.ID,
			TenantID: user.TenantID,
		})
	}

	if err := s.tokenRepo.Create(tokenCtx, resetToken); err != nil {
		return err
	}

	if err := s.mailer.SendPasswordReset(context.Background(), user.Email.String(), resetToken.Token); err != nil {
		log.Printf("Falha ao enfileirar e-mail de redefinição de senha: %v", err)
	}

	return nil
}

// ResetPassword validates the token, updates the password and revokes all active sessions in an atomic transaction.
func (s *authService) ResetPassword(ctx context.Context, rawToken string, rawNewPassword string) error {
	resetToken, err := s.tokenRepo.FindByToken(ctx, rawToken)
	if err != nil || resetToken == nil {
		return ErrResetTokenInvalid
	}

	if time.Now().UTC().After(resetToken.ExpiresAt) {
		_ = s.tokenRepo.Delete(ctx, rawToken)
		return ErrResetTokenExpired
	}

	newPassword, err := pkgDomain.NewPassword(rawNewPassword)
	if err != nil {
		return err
	}

	txCtxScoped := ctx
	if ctx.Value(token.ClaimsContextKey) == nil && resetToken.TenantID != uuid.Nil {
		txCtxScoped = context.WithValue(ctx, token.ClaimsContextKey, &token.CustomClaims{
			UserID:   resetToken.UserID,
			TenantID: resetToken.TenantID,
		})
	}

	user, err := s.userRepo.FindByID(txCtxScoped, resetToken.UserID)
	if err != nil || user == nil {
		return ErrResetTokenInvalid
	}

	user.Password = newPassword

	resetFn := func(txCtx context.Context) error {
		if err := s.userRepo.Update(txCtx, user); err != nil {
			return err
		}
		// Invalida o token de reset usado e revoga os refresh tokens existentes do usuário
		_ = s.tokenRepo.Delete(txCtx, rawToken)
		return s.tokenRepo.DeleteByUserID(txCtx, resetToken.UserID)
	}

	if s.tm != nil {
		if err := s.tm.Transaction(txCtxScoped, resetFn); err != nil {
			return err
		}
	} else {
		if err := resetFn(txCtxScoped); err != nil {
			return err
		}
	}

	return nil
}

func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, dto domain.ChangePasswordDTO) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return ErrInvalidCredentials
	}

	if err := user.Password.Compare(dto.CurrentPassword); err != nil {
		return ErrCurrentPasswordIncorrect
	}

	newPassword, err := pkgDomain.NewPassword(dto.NewPassword)
	if err != nil {
		return err
	}

	user.Password = newPassword
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	_ = s.tokenRepo.DeleteByUserID(ctx, userID)

	return nil
}
