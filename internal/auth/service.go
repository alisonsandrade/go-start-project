// Package auth service
package auth

import (
	"strconv"
	"strings"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/auth/domain"
	"github.com/alisonsandrade/go-start-project/internal/config"
	usersDomain "github.com/alisonsandrade/go-start-project/internal/users/domain"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(user *usersDomain.User) error
	FindByEmail(email string) (*usersDomain.User, error)
	FindByID(id uuid.UUID) (*usersDomain.User, error)
}

type AuthService interface {
	Register(dto domain.RegisterRequest) (*domain.AuthResponseDTO, error)
	Login(dto domain.LoginRequest) (*domain.AuthResponseDTO, error)
	RefreshSession(refreshToken string) (*domain.AuthResponseDTO, error)
	Logout(userID uuid.UUID) error
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
func (s *authService) Register(dto domain.RegisterRequest) (*domain.AuthResponseDTO, error) {
	existingUser, err := s.userRepo.FindByEmail(dto.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	user := &usersDomain.User{
		Name:      dto.Name,
		Email:     strings.ToLower(strings.TrimSpace(dto.Email)),
		Password:  dto.Password,
		Role:      usersDomain.RoleUser,
		Phone:     dto.Phone,
		AvatarURL: dto.AvatarURL,
		JobTitle:  dto.JobTitle,
		Bio:       dto.Bio,
		IsActive:  true,
	}

	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(user)
}

// Login authenticates the given credentials and returns a new token pair.
func (s *authService) Login(dto domain.LoginRequest) (*domain.AuthResponseDTO, error) {
	user, err := s.userRepo.FindByEmail(dto.Email)
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}
	if !user.CheckPassword(dto.Password) {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	return s.generateAuthResponse(user)
}

// Logout invalidates the user's session by deleting all of their refresh tokens.
func (s *authService) Logout(userID uuid.UUID) error {
	return s.tokenRepo.DeleteByUserID(userID)
}

// RefreshSession validates the refresh token, rotates it, and returns a new token pair.
func (s *authService) RefreshSession(refreshToken string) (*domain.AuthResponseDTO, error) {
	rt, err := s.tokenRepo.FindByToken(refreshToken)
	if err != nil || rt == nil {
		return nil, ErrInvalidRefreshToken
	}

	if time.Now().After(rt.ExpiresAt) {
		_ = s.tokenRepo.Delete(refreshToken)
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.FindByID(rt.UserID)
	if err != nil || user == nil {
		return nil, ErrInvalidRefreshToken
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Rotation invalid old refresh token
	if err := s.tokenRepo.Delete(refreshToken); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(user)
}

// generateAuthResponse builds the access and refresh tokens and persists the refresh token.
func (s *authService) generateAuthResponse(user *usersDomain.User) (*domain.AuthResponseDTO, error) {
	expHours, _ := strconv.Atoi(s.cfg.JWTExpirationHours)
	if expHours == 0 {
		expHours = 24
	}

	accessToken, err := token.GenerateToken(
		user.ID, user.Email, string(user.Role), s.cfg.JWTSecret, expHours,
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
	if err := s.tokenRepo.Create(rt); err != nil {
		return nil, err
	}

	return &domain.AuthResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		User:         *user,
	}, nil
}
