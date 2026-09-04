// Package users service onde estão todos os serviços e referências ao repository da entidade users
package users

import (
	"context"
	"errors"

	"github.com/alisonsandrade/go-start-project/internal/roles"
	"github.com/alisonsandrade/go-start-project/internal/users/domain"
	pkgDomain "github.com/alisonsandrade/go-start-project/pkg/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrUserNotFound is returned when a user cannot be found.
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidRole is returned when an invalid role is provided.
	ErrInvalidRole = errors.New("invalid role")

	// ErrEmailAlreadyExists indicates the email is already registered.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrForbidden indicates not user permission
	ErrForbidden = errors.New("permissão insuficiente: requer perfil admin")
)

// UserService defines the contract for user-related business logic.
type UserService interface {
	GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	GetDefaultRoleID(ctx context.Context) (uuid.UUID, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, dto domain.UpdateUserRequest) (*domain.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error

	// --- Admin-exclusive methods ---
	ListUsers(ctx context.Context, params pagination.Params) (pagination.PageResult[domain.User], error)
	CreateUserAsAdmin(ctx context.Context, dto domain.CreateUserRequest) (*domain.User, error)
	UpdateUserAsAdmin(ctx context.Context, id uuid.UUID, dto domain.AdminUpdateUserRequest) error
	SoftDeleteUserAsAdmin(ctx context.Context, id uuid.UUID) error

	// Seed
	SeedDefaultAdmin(ctx context.Context, name, email, password string) error
}

type userService struct {
	userRepo UserRepository
	roleRepo roles.RoleRepository
}

func NewUserService(userRepo UserRepository, roleRepo roles.RoleRepository) UserService {
	return &userService{userRepo: userRepo, roleRepo: roleRepo}
}

/*************************************************************************************************
*								Services' Users
*************************************************************************************************/
func (s *userService) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) GetDefaultRoleID(ctx context.Context) (uuid.UUID, error) {
	return s.userRepo.GetDefaultRoleID(ctx)
}

func (s *userService) UpdateUser(ctx context.Context, userID uuid.UUID, dto domain.UpdateUserRequest) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
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

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, userID)
}

func (s *userService) ListUsers(
	ctx context.Context,
	params pagination.Params,
) (pagination.PageResult[domain.User], error) {
	users, total, err := s.userRepo.List(ctx, params.Limit, params.Offset())
	if err != nil {
		return pagination.PageResult[domain.User]{}, err
	}

	return pagination.NewPageResult(users, total, params), nil
}

/*************************************************************************************************
*							Services' Users for Admin
*************************************************************************************************/
func (s *userService) CreateUserAsAdmin(
	ctx context.Context,
	userDTO domain.CreateUserRequest,
) (*domain.User, error) {
	email, err := pkgDomain.NewEmail(userDTO.Email)
	if err != nil {
		return nil, err
	}

	_, err = s.roleRepo.GetByID(userDTO.RoleID)
	if err != nil {
		return nil, ErrInvalidRole
	}

	existingUser, err := s.userRepo.FindByEmail(ctx, email.String())
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	password, err := pkgDomain.NewPassword(userDTO.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:      userDTO.Name,
		Email:     email,
		Password:  password,
		RoleID:    userDTO.RoleID,
		Phone:     userDTO.Phone,
		AvatarURL: userDTO.AvatarURL,
		JobTitle:  userDTO.JobTitle,
		Bio:       userDTO.Bio,
		IsActive:  true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) UpdateUserAsAdmin(ctx context.Context, id uuid.UUID, dto domain.AdminUpdateUserRequest) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	if dto.Name != nil {
		user.Name = *dto.Name
	}

	if dto.Email != nil {
		if *dto.Email != user.Email.String() {
			newEmail, err := pkgDomain.NewEmail(*dto.Email) // use o alias do seu pacote pkg/domain
			if err != nil {
				if errors.Is(err, pkgDomain.ErrInvalidEmail) {
					return pkgDomain.ErrInvalidEmail
				}
				return err
			}

			existingUser, _ := s.userRepo.FindByEmail(ctx, newEmail.String())
			if existingUser != nil {
				return ErrEmailAlreadyExists
			}

			user.Email = newEmail
		}
	}

	if dto.IsActive != nil {
		user.IsActive = *dto.IsActive
	}

	return s.userRepo.Update(ctx, user)
}

func (s *userService) SoftDeleteUserAsAdmin(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}
