package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestTenantValidate(t *testing.T) {
	t.Run("normalizes document and initializes settings", func(t *testing.T) {
		tenant := &Tenant{Name: "  Example Ltda  ", Document: "12.345.678/0001-90"}

		require.NoError(t, tenant.Validate())
		assert.Equal(t, "Example Ltda", tenant.Name)
		assert.Equal(t, "12345678000190", tenant.Document)
		assert.JSONEq(t, `{}`, string(tenant.Settings))
	})

	t.Run("rejects missing name", func(t *testing.T) {
		err := (&Tenant{Document: "12345"}).Validate()
		assert.ErrorIs(t, err, ErrNameRequired)
	})

	t.Run("rejects missing document", func(t *testing.T) {
		err := (&Tenant{Name: "Example"}).Validate()
		assert.ErrorIs(t, err, ErrDocumentRequired)
	})

	t.Run("rejects invalid document length", func(t *testing.T) {
		err := (&Tenant{Name: "Example", Document: "1234"}).Validate()
		assert.ErrorIs(t, err, ErrDocumentInvalid)
	})

	t.Run("rejects invalid settings", func(t *testing.T) {
		err := (&Tenant{Name: "Example", Document: "12345", Settings: datatypes.JSON(`{invalid`)}).Validate()
		assert.ErrorIs(t, err, ErrSettingsInvalid)
	})
}
