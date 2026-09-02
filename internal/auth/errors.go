// Package auth
package auth

import "errors"

// Domain-level errors returned by the service layer.
// Handlers map these to HTTP status codes; repositories never know about them.
var (
	// ErrCurrentPasswordIncorrect indicates when then password is incorrect
	ErrCurrentPasswordIncorrect = errors.New("current password is incorrect")

	// ErrEmailAlreadyExists indicates the email is already registered.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrInvalidCredentials indicates the provided credentials are incorrect.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserInactive indicates the user account is deactivated.
	ErrUserInactive = errors.New("user is inactive")

	// ErrInvalidRefreshToken indicates the refresh token is invalid or expired.
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")

	// ErrResetTokenInvalid indicates when the token is invalid
	ErrResetTokenInvalid = errors.New("reset token is invalid or does not exist")

	// ErrResetTokenExpired indicates when the token has expired
	ErrResetTokenExpired = errors.New("reset token has expired")

	// ErrResetTokenUsed indicates when token already been used
	ErrResetTokenUsed = errors.New("reset token has already been used")
)
