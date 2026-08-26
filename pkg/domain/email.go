package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
)

var (
	ErrEmptyEmail   = errors.New("email cannot be empty")
	ErrInvalidEmail = errors.New("invalid email format")
)

// Email represents an value object for an email address.
type Email struct {
	value string
}

// NewEmail valid, sanitise and instance a Value Object
func NewEmail(raw string) (Email, error) {
	trimmed := strings.ToLower(strings.TrimSpace(raw))

	if trimmed == "" {
		return Email{}, ErrEmptyEmail
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil || addr.Address != trimmed {
		return Email{}, ErrInvalidEmail
	}

	parts := strings.Split(trimmed, "@")
	if len(parts) != 2 || !strings.Contains(parts[1], ".") {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: trimmed}, nil
}

// String returns the primitive value of the Email Value Object
func (e Email) String() string {
	return e.value
}

// Interfaces para o GORM e o driver SQL lerem/gravarem diretamente como string
func (e Email) Value() (driver.Value, error) {
	return e.value, nil
}

func (e *Email) Scan(value interface{}) error {
	if value == nil {
		e.value = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		e.value = v
	case []byte:
		e.value = string(v)
	default:
		return fmt.Errorf("tipo incompatível para Email: %T", value)
	}
	return nil
}

// Interfaces para serialização/deserialização JSON
func (e Email) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.value)
}

func (e *Email) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	email, err := NewEmail(raw)
	if err != nil {
		return err
	}
	*e = email
	return nil
}
