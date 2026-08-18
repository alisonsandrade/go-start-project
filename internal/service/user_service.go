// Package service onde estão todos os serviços e referências ao repository da entidade users
package service

import (
	"errors"

	"github.com/alisonsandrade/go-start-project/internal/domain"
	"github.com/alisonsandrade/go-start-project/internal/repository"
	"github.com/google/uuid"
)

var (
	// ErrUserNotFound is returned when a user cannot be found.
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidRole is returned when an invalid role is provided.
	ErrInvalidRole = errors.New("invalid role")
)

// UserService defines the contract for user-related business logic.
type UserService interface {
	GetUser(userID uuid.UUID) (*domain.User, error)
	UpdateUser(userID uuid.UUID, dto domain.UpdateUserDTO) (*domain.User, error)
	DeleteUser(userID uuid.UUID) error

	// --- Admin-exclusive methods ---
	ListUsers() ([]domain.User, error)
	GetUserByID(id uuid.UUID) (*domain.User, error)
	UpdateUserAsAdmin(id uuid.UUID, dto domain.AdminUpdateUserDTO) error
	DeleteUserAsAdmin(id uuid.UUID) error
}

type userService struct{ userRepo repository.UserRepository }

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

/*************************************************************************************************
*								Services' Users
*************************************************************************************************/
func (s *userService) GetUser(userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) UpdateUser(userID uuid.UUID, dto domain.UpdateUserDTO) (*domain.User, error) {
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

func (s *userService) ListUsers() ([]domain.User, error) {
	return s.userRepo.ListAll()
}

/*************************************************************************************************
*							Services' Users for Admin
*************************************************************************************************/
func (s *userService) GetUserByID(id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) UpdateUserAsAdmin(id uuid.UUID, dto domain.AdminUpdateUserDTO) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	if dto.Name != nil {
		user.Name = *dto.Name
	}

	if dto.Email != nil {
		if *dto.Email != user.Email {
			existingUser, _ := s.userRepo.FindByEmail(*dto.Email)
			if existingUser != nil {
				return ErrEmailAlreadyExists
			}
			user.Email = *dto.Email
		}
	}

	if dto.Role != nil {
		roleStr := domain.Role(*dto.Role)
		if roleStr != domain.RoleUser && roleStr != domain.RoleAdmin {
			return ErrInvalidRole
		}
		user.Role = roleStr
	}

	if dto.IsActive != nil {
		user.IsActive = *dto.IsActive
	}

	return s.userRepo.Update(user)
}

func (s *userService) DeleteUserAsAdmin(id uuid.UUID) error {
	return s.userRepo.Delete(id)
}
