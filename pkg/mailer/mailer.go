// Package mailer provides a centralized service for sending asynchronous emails.
package mailer

import "context"

// Mailer defines the methods available for sending emails.
// Any module the needs to send emails should depend on this interface.
type Mailer interface {
	SendPasswordReset(ctx context.Context, toEmail, token string) error
	Close() error
}
