// Package domain_test
package domain_test

import (
	"testing"

	"github.com/alisonsandrade/go-start-project/pkg/domain"
)

func TestEmailValueObject(t *testing.T) {
	t.Run("sucesso ao criar email valido com espaços ou maiusculas", func(t *testing.T) {
		email, err := domain.NewEmail("  USER@Dominio.COM  ")
		if err != nil {
			t.Fatalf("esperava erro nil, recebeu: %v", err)
		}
		if email.String() != "user@dominio.com" {
			t.Errorf("esperava 'user@dominio.com', obteve '%s'", email.String())
		}
	})

	t.Run("falha ao enviar email invalido", func(t *testing.T) {
		invalidCases := []string{"", "user@", "user@dominio", "@dominio.com", "plainaddress"}
		for _, raw := range invalidCases {
			_, err := domain.NewEmail(raw)
			if err == nil {
				t.Errorf("esperava erro para email '%s', mas obteve nil", raw)
			}
		}
	})
}

func TestPasswordValueObject(t *testing.T) {
	t.Run("sucesso ao criar senha forte e verificar hash", func(t *testing.T) {
		raw := "Senha@Forte123"
		pass, err := domain.NewPassword(raw)
		if err != nil {
			t.Fatalf("esperava erro nil, obteve: %v", err)
		}

		if err := pass.Compare(raw); err != nil {
			t.Errorf("esperava que a senha validasse contra o próprio hash, erro: %v", err)
		}

		if err := pass.Compare("SenhaErrada123"); err == nil {
			t.Errorf("nao deveria validar hash com senha incorreta")
		}
	})

	t.Run("falha ao criar senha fraca", func(t *testing.T) {
		weakPasswords := []string{
			"curta",         // menor que 8 caracteres
			"somenteletras", // sem número
		}
		for _, raw := range weakPasswords {
			_, err := domain.NewPassword(raw)
			if err == nil {
				t.Errorf("esperava erro para a senha fraca '%s', mas obteve nil", raw)
			}
		}
	})
}
