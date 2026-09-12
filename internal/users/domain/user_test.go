package domain

import (
	"strings"
	"testing"

	"github.com/alisonsandrade/go-start-project/pkg/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	roleID := uuid.New()

	user, err := NewUser("  Alice Smith  ", "ALICE@example.com", "StrongPass1", roleID)

	require.NoError(t, err)
	assert.Equal(t, "Alice Smith", user.Name)
	assert.Equal(t, "alice@example.com", user.Email.String())
	assert.Equal(t, roleID, user.RoleID)
	assert.True(t, user.IsActive)
	assert.NotEmpty(t, user.Password.Hash())
}

func TestNewUserRejectsInvalidValues(t *testing.T) {
	roleID := uuid.New()

	tests := []struct {
		name     string
		email    string
		password string
		expected error
	}{
		{name: "invalid email", email: "invalid", password: "StrongPass1", expected: domain.ErrInvalidEmail},
		{name: "weak password", email: "user@example.com", password: "short", expected: domain.ErrPasswordTooShort},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser("Alice", tt.email, tt.password, roleID)

			assert.Nil(t, user)
			assert.ErrorIs(t, err, tt.expected)
		})
	}
}

func TestUserValidate(t *testing.T) {
	roleID := uuid.New()
	validUser := func() *User {
		return &User{Name: "Alice", RoleID: roleID}
	}

	tests := []struct {
		name     string
		mutate   func(*User)
		expected error
	}{
		{name: "trims mutable text fields", mutate: func(user *User) {
			user.Name = "  Alice  "
			user.Phone = " 123 "
			user.JobTitle = " Engineer "
			user.AvatarURL = " https://example.com/avatar.jpg "
		}, expected: nil},
		{name: "requires a name with at least three characters", mutate: func(user *User) { user.Name = "Al" }, expected: ErrUserNameTooShort},
		{name: "limits the name length", mutate: func(user *User) { user.Name = strings.Repeat("a", 101) }, expected: ErrUserNameTooLong},
		{name: "requires a role", mutate: func(user *User) { user.RoleID = uuid.Nil }, expected: ErrUserRoleRequired},
		{name: "limits the phone length", mutate: func(user *User) { user.Phone = strings.Repeat("1", 21) }, expected: ErrUserPhoneTooLong},
		{name: "limits the job title length", mutate: func(user *User) { user.JobTitle = strings.Repeat("a", 101) }, expected: ErrUserJobTitleLong},
		{name: "limits the avatar URL length", mutate: func(user *User) { user.AvatarURL = strings.Repeat("a", 256) }, expected: ErrUserAvatarTooLong},
		{name: "requires a valid avatar URL", mutate: func(user *User) { user.AvatarURL = "://invalid" }, expected: ErrUserAvatarInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := validUser()
			tt.mutate(user)

			err := user.Validate()

			assert.ErrorIs(t, err, tt.expected)
			if tt.expected == nil {
				assert.Equal(t, "Alice", user.Name)
				assert.Equal(t, "123", user.Phone)
				assert.Equal(t, "Engineer", user.JobTitle)
			}
		})
	}
}

func TestUserBeforeCreateAssignsID(t *testing.T) {
	user := &User{}

	require.NoError(t, user.BeforeCreate(nil))

	assert.NotEqual(t, uuid.Nil, user.ID)
}
