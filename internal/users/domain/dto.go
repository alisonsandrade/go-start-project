// Package domain users dto.
package domain

// UserBase for interface
type UserBase struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	AvatarURL string `json:"avatar_url"`
	JobTitle  string `json:"job_title"`
	Bio       string `json:"bio"`
}

type UpdateUserRequest struct {
	Name      string `json:"name,omitempty" example:"Alison Andrade"`
	Phone     string `json:"phone,omitempty" example:"7599999999"`
	AvatarURL string `json:"avatar_url,omitempty" example:"https://github.com/alisonsandrade/avatar.jpg"`
	JobTitle  string `json:"job_title,omitempty" example:"Tech Lead"`
	Bio       string `json:"bio,omitempty" example:"Desenvolvedor Backend Go"`
}

// AdminCreateUserRequest represents the payload for an admin to create any user's data.
type CreateUserRequest struct {
	UserBase
	Password string `json:"password"`
	Role     Role   `json:"role"`
}

// AdminCreateUserResponse represents the response admin to create new user.
type CreateUserResponse struct {
	UserBase
	Role Role `json:"role"`
}

// AdminUpdateUserRequest represents the payload for an admin to update any user's data.
// Pointers are used to allow partial updates (PATCH behavior).
type AdminUpdateUserRequest struct {
	Name     *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
	Role     *Role   `json:"role,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}
