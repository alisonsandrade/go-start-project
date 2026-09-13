// Package domain - RBAC rules (roles and permissions).
package domain

import (
	"strings"

	"github.com/alisonsandrade/go-start-project/internal/domain"
)

// Permission represents a fine-grained action, persisted in the database
type Permission struct {
	domain.BaseModel `swaggerignore:"true"`
	Code             string `gorm:"size:100;uniqueIndex;not null" json:"code"`
	Description      string `gorm:"size:255" json:"description"`
}

// RoleEntity represents a role (a named set of permissions) persisted in the database
type RoleEntity struct {
	domain.BaseModelTenant `swaggerignore:"true"`
	Name                   string       `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Description            string       `gorm:"size:255" json:"description"`
	IsSystem               bool         `gorm:"not null;default:false" json:"is_system"`
	Permissions            []Permission `gorm:"many2many:role_permissions;joinForeignKey:RoleID;joinReferences:PermissionID" json:"permissions"`
}

func (RoleEntity) TableName() string {
	return "roles"
}

func (Permission) TableName() string {
	return "permissions"
}

func NormalizeRoleName(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}

type PermissionCode string

const (
	// User permissions
	PermissionReadUser   PermissionCode = "user:read"
	PermissionCreateUser PermissionCode = "user:create"
	PermissionUpdateUser PermissionCode = "user:update"
	PermissionDeleteUser PermissionCode = "user:delete"
	PermissionListUsers  PermissionCode = "user:list"

	// Role permissions
	PermissionReadRole   PermissionCode = "role:read"
	PermissionCreateRole PermissionCode = "role:create"
	PermissionUpdateRole PermissionCode = "role:update"
	PermissionDeleteRole PermissionCode = "role:delete"

	// Permission assignment
	PermissionAssignRolePermissions PermissionCode = "role:assign-permissions"

	// Tenant permissions
	PermissionManageTenant PermissionCode = "tenant:manage"
)
