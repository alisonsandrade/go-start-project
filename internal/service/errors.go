// Package service
package service

import "errors"

// Domain-level errors returned by the service layer.
// Handlers map these to HTTP status codes; repositories never know about them.
var (
	// ErrRoleNotFound is returned when a role cannot be found by ID or name.
	ErrRoleNotFound = errors.New("role not found")

	// ErrRoleAlreadyExists is returned when creating a role whose name is taken.
	ErrRoleAlreadyExists = errors.New("role already exists")

	// ErrSystemRoleImmutable is returned when trying to modify or delete a
	// system role (is_system = true), such as ADMIN or USER.
	ErrSystemRoleImmutable = errors.New("system roles cannot be modified or deleted")

	// ErrInvalidPermissions is returned when one or more permission IDs do not exist.
	ErrInvalidPermissions = errors.New("one or more permissions do not exist")
)
