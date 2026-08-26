package domain

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordTooShort = errors.New("password should be at least 8 characters")
	ErrPasswordNoNumber = errors.New("password should contain at least one number")
	ErrPasswordMismatch = errors.New("password mismatch")
)

type Password struct {
	hash string
}

// NewPassword validates complexity rules and generates a secure hash
func NewPassword(plainText string) (Password, error) {
	if len(plainText) < 8 {
		return Password{}, ErrPasswordTooShort
	}

	var hasNumber bool
	for _, char := range plainText {
		if unicode.IsDigit(char) {
			hasNumber = true
			break
		}
	}

	if !hasNumber {
		return Password{}, ErrPasswordNoNumber
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainText), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, err
	}

	return Password{hash: string(hashedBytes)}, nil
}

// PasswordFromHash builds the Value Object from hash without validating complexity rules.
func PasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

// Compare verifies if the password matches the stored hash verifies if the password matches the stored hash
func (p Password) Compare(plainText string) error {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plainText))
	if err != nil {
		return ErrPasswordMismatch
	}
	return nil
}

// Hash returns the hash ready for persistence in the database
func (p Password) Hash() string {
	return p.hash
}

// Hidden completily the password in JSON
func (p Password) MarshalJSON() ([]byte, error) {
	return []byte(`""`), nil
}

// Interfaces for GORM/SQL to persist the hash
func (p Password) Value() (driver.Value, error) {
	return p.hash, nil
}

func (p *Password) Scan(value interface{}) error {
	if value == nil {
		p.hash = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		p.hash = v
	case []byte:
		p.hash = string(v)
	default:
		return fmt.Errorf("tipo incompatível para Password: %T", value)
	}
	return nil
}
