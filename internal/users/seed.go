package users

import "context"

// SeedDefaultAdmin cria o admin inicial se ele ainda não existir na base
func (s *userService) SeedDefaultAdmin(ctx context.Context, name, rawEmail, rawPassword string) error {
	/*
		email, err := pkgDomain.NewEmail(rawEmail)
		if err != nil {
			return fmt.Errorf("e-mail padrão de admin inválido: %w", err)
		}

		existingUser, err := s.userRepo.FindByEmail(ctx, email.String())
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("erro ao verificar existência do admin: %w", err)
		}
		if existingUser != nil {
			return nil
		}

		password, err := pkgDomain.NewPassword(rawPassword)
		if err != nil {
			return fmt.Errorf("senha padrão de admin inválida: %w", err)
		}

		adminUser := &usersDomain.User{
			Name:     name,
			Email:    email,
			Password: password,
			RoleID:   usersDomain.RoleAdmin,
			IsActive: true,
		}

		if err := s.userRepo.Create(ctx, adminUser); err != nil {
			return fmt.Errorf("falha ao persistir admin padrão: %w", err)
		}

		log.Printf("✔ Usuário administrador padrão criado com sucesso: %s", email.String())

	*/
	return nil
}
