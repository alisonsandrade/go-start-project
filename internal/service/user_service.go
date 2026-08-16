package service

import (
	"errors"
	"strconv"

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
)

type UserService interface {
	Register(dto domain.CreateUserDTO) (*domain.User, error)
	UpdateProfile(userID uuid.UUID, dto domain.UpdateUserDTO) (*domain.User, error)
	Login(dto domain.LoginDTO) (*domain.AuthResponseDTO, error)
	GetProfile(userID uuid.UUID) (*domain.User, error)
	ListUsers() ([]domain.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

func NewUserService(userRepo repository.UserRepository, cfg *config.Config) UserService {
	return &userService{
		userRepo: userRepo,
		cfg:      cfg,
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

	role := dto.Role
	if role == "" {
		role = domain.RoleUser
	}

	user := &domain.User{
		Name:      dto.Name,
		Email:     dto.Email,
		Password:  dto.Password,
		Role:      role,
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

func (s *userService) Login(dto domain.LoginDTO) (*domain.AuthResponseDTO, error) {
	user, err := s.userRepo.FindByEmail(dto.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	if !user.CheckPassword(dto.Password) {
		return nil, ErrInvalidCredentials
	}
	expHours, _ := strconv.Atoi(s.cfg.JWTExpirationHours)
	if expHours == 0 {
		expHours = 24
	}

	tokenString, err := token.GenerateToken(
		user.ID,
		user.Email,
		string(user.Role),
		s.cfg.JWTSecret,
		expHours,
	)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponseDTO{
		Token: tokenString,
		User:  *user,
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
