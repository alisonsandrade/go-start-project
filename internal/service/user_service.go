// Package service onde estão todos os serviços e referências ao repository da entidade users
package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/domain"
	"github.com/alisonsandrade/go-start-project/internal/repository"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
)

var (
	ErrEmailAlreadyExists = errors.New("e-mail já cadastrado no sistema")
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrUserNotFound       = errors.New("usuário não encontrado")
	ErrInvalidRole        = errors.New("perfil (role) inválido")
)

type UserService interface {
	Register(dto domain.CreateUserDTO) (*domain.User, error)
	UpdateProfile(userID uuid.UUID, dto domain.UpdateUserDTO) (*domain.User, error)
	DeleteUser(userID uuid.UUID) error
	Login(dto domain.LoginDTO) (*domain.AuthResponseDTO, error)
	RefreshSession(refreshToken string) (*domain.AuthResponseDTO, error)
	GetProfile(userID uuid.UUID) (*domain.User, error)
	ListUsers() ([]domain.User, error)
}

type userService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	cfg       *config.Config
}

func NewUserService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	cfg *config.Config,
) UserService {
	return &userService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
	}
}

func (s *userService) Register(dto domain.CreateUserDTO) (*domain.User, error) {
	existingUser, err := s.userRepo.FindByEmail(dto.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	role := strings.ToUpper(strings.TrimSpace(string(dto.Role)))
	if role == "" {
		role = string(domain.RoleUser)
	}
	if role != string(domain.RoleUser) && role != string(domain.RoleAdmin) {
		return nil, ErrInvalidRole
	}

	user := &domain.User{
		Name:      dto.Name,
		Email:     dto.Email,
		Password:  dto.Password,
		Role:      domain.Role(role),
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

	return user, nil
}

func (s *userService) UpdateProfile(userID uuid.UUID, dto domain.UpdateUserDTO) (*domain.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Validação dos campos que serão alterados
	if dto.Name != "" {
		user.Name = dto.Name
	}
	if dto.Phone != "" {
		user.Phone = dto.Phone
	}
	if dto.AvatarURL != "" {
		user.AvatarURL = dto.AvatarURL
	}
	if dto.JobTitle != "" {
		user.JobTitle = dto.JobTitle
	}
	if dto.Bio != "" {
		user.Bio = dto.Bio
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) DeleteUser(userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(userID)
}

func (s *userService) Login(dto domain.LoginDTO) (*domain.AuthResponseDTO, error) {
	user, err := s.userRepo.FindByEmail(dto.Email)
	if err != nil || user == nil {
		return nil, errors.New("credenciais inválidas")
	}
	if !user.CheckPassword(dto.Password) {
		return nil, ErrInvalidCredentials
	}
	expHours, _ := strconv.Atoi(s.cfg.JWTExpirationHours)
	if expHours == 0 {
		expHours = 24
	}

	// token generate
	accessToken, err := token.GenerateToken(
		user.ID,
		user.Email,
		string(user.Role),
		s.cfg.JWTSecret,
		expHours,
	)
	if err != nil {
		return nil, err
	}

	// refresh token generate
	refreshTokenStr, err := token.GenerateSecureToken()
	if err != nil {
		return nil, errors.New("erro ao gear o token de renovação")
	}
	rt := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	if err := s.tokenRepo.Create(rt); err != nil {
		return nil, errors.New("erro ao salvar sessão no banco de dados")
	}

	return &domain.AuthResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		User:         *user,
	}, nil
}

func (s *userService) GetProfile(userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) ListUsers() ([]domain.User, error) {
	return s.userRepo.ListAll()
}

func (s *userService) RefreshSession(refreshToken string) (*domain.AuthResponseDTO, error) {
	// 1. Busca o refresh token no banco de dados
	rt, err := s.tokenRepo.FindByToken(refreshToken)
	if err != nil || rt == nil {
		return nil, errors.New("token de renovação inválido ou não encontrado")
	}

	// 2. Verifica se o refresh token já passou da validade
	if time.Now().After(rt.ExpiresAt) {
		_ = s.tokenRepo.Delete(refreshToken)
		return nil, errors.New("sessão expirada, por favor faça login novamente")
	}

	// 3. Resgata o usuário associado para preencher o novo JWT
	user, err := s.userRepo.FindByID(rt.UserID)
	if err != nil || user == nil {
		return nil, errors.New("usuário associado não encontrado")
	}

	// 4. Gera um NOVO Access Token
	newAccessToken, err := token.GenerateToken(user.ID, user.Email, string(user.Role), s.cfg.JWTSecret, 1)
	if err != nil {
		return nil, err
	}

	// 5. Gera um NOVO Refresh Token (Rotatividade de Segurança)
	newRefreshTokenStr, err := token.GenerateSecureToken()
	if err != nil {
		return nil, err
	}

	// 6. Atualiza o banco de dados (Deleta o velho e salva o novo)
	_ = s.tokenRepo.Delete(refreshToken)

	newRT := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     newRefreshTokenStr,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	if err := s.tokenRepo.Create(newRT); err != nil {
		return nil, errors.New("erro ao atualizar sessão no banco de dados")
	}

	return &domain.AuthResponseDTO{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenStr,
		User:         *user,
	}, nil
}
