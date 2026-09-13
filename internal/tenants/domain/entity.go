package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode"

	baseDomain "github.com/alisonsandrade/go-start-project/internal/domain"
	"gorm.io/datatypes"
)

var (
	ErrNameRequired     = errors.New("tenant name is required")
	ErrDocumentRequired = errors.New("tenant document is required")
	ErrDocumentInvalid  = errors.New("tenant document is invalid")
	ErrSettingsInvalid  = errors.New("tenant settings must be valid JSON")
)

type Tenant struct {
	baseDomain.BaseModel `swaggerignore:"true"`
	Name                 string         `gorm:"size:150;not null" json:"name"`
	Document             string         `gorm:"size:30;not null;uniqueIndex" json:"document"`
	Settings             datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"settings" swaggertype:"object"`
}

func (t *Tenant) Normalize() {
	t.Name = strings.TrimSpace(t.Name)
	t.Document = normalizeDocument(t.Document)
	if len(t.Settings) == 0 {
		t.Settings = datatypes.JSON([]byte(`{}`))
	}
}

func (t *Tenant) Validate() error {
	t.Normalize()
	if t.Name == "" {
		return ErrNameRequired
	}
	if !json.Valid(t.Settings) {
		return ErrSettingsInvalid
	}
	if t.Document == "" {
		return ErrDocumentRequired
	}
	if len(t.Document) < 5 || len(t.Document) > 30 {
		return ErrDocumentInvalid
	}
	return nil
}

func normalizeDocument(document string) string {
	var normalized strings.Builder
	for _, character := range strings.TrimSpace(document) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			normalized.WriteRune(unicode.ToUpper(character))
		}
	}
	return normalized.String()
}

func (Tenant) TableName() string {
	return "tenants"
}
