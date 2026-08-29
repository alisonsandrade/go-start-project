// Package auth service
package auth

import (
	"context"
	"strconv"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/auth/domain"
	"github.com/alisonsandrade/go-start-project/internal/config"
	usersDomain "github.com/alisonsandrade/go-start-project/internal/users/domain"
	pkgDomain "github.com/alisonsandrade/go-start-project/pkg/domain"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, dto domain.RegisterRequest) (*domain.AuthResponseDTO, error)
	Login(ctx context.Context, dto domain.LoginRequest) (*domain.AuthResponseDTO, error)
	RefreshSession(ctx context.Context, refreshToken string) (*domain.AuthResponseDTO, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}

type authService struct {
	userRepo  UserRepository
	tokenRepo TokenRepository
	cfg       *config.Config
}

func NewAuthService(
	userRepo UserRepository,
	tokenRepo TokenRepository,
	cfg *config.Config,
) AuthService {
	return &authService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
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

	user := &usersDomain.User{
		Name:     dto.Name,
		Email:    email,
		Password: password,
		// Role:      usersDomain.RoleUser,
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
