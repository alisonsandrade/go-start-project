// Package app onde estão centralizados todos os roteamentos e ciclo de vida do aplicativo
package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	_ "github.com/alisonsandrade/go-start-project/docs"
	"github.com/alisonsandrade/go-start-project/pkg/mailer"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gorm.io/gorm"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/platform/database"
	"github.com/alisonsandrade/go-start-project/internal/roles"
	"github.com/alisonsandrade/go-start-project/internal/users"
)

type App struct {
	Config *config.Config
	DB     *gorm.DB
	Router *chi.Mux
	Mailer mailer.Mailer
}

func New() (*App, error) {
	// 1. Carrega configurações
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("falha ao carregar configs: %w", err)
	}

	// 2. Conecta ao Banco de Dados
	db, err := database.NewDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no banco: %w", err)
	}

	// 3. Inicializa Roteador e Middlewares Globais
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// Inicializa o serviço de e-mail
	emailService := mailer.NewWorkerMailer(3, 100)

	// 4. Dependency wiring
	userRepo := users.NewUserRepository(db)
	tokenRepo := auth.NewTokenRepository(db)
	roleRepo := roles.NewRoleRepository(db)

	authService := auth.NewAuthService(userRepo, tokenRepo, cfg, emailService)
	userService := users.NewUserService(userRepo, roleRepo)
	roleService := roles.NewRoleService(roleRepo)

	// --- SEED DO ADMIN INICIAL ---
	ctx := context.Background()
	if err := userService.SeedDefaultAdmin(
		ctx,
		cfg.AdminSeed.Name,
		cfg.AdminSeed.Email,
		cfg.AdminSeed.Password,
	); err != nil {
		log.Printf("⚠ Aviso: não foi possível executar a seed do admin: %v", err)
	}

	authHandler := auth.NewAuthHandler(authService)
	userHandler := users.NewUserHandler(userService)
	roleHandler := roles.NewRoleHandler(roleService)

	// 5. Route registration by domain
	r.Route("/api", func(api chi.Router) {
		api.Mount("/auth", authHandler.AuthRoutes(cfg))
		api.Mount("/users", userHandler.Routes(cfg, roleRepo))
		api.Mount("/roles", roleHandler.Routes(cfg, roleRepo))
	})

	// Documentação Swagger
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%s/swagger/doc.json", cfg.Port)),
	))

	return &App{
		Config: cfg,
		DB:     db,
		Router: r,
		Mailer: emailService,
	}, nil
}

func (a *App) Stop() {
	if a.Mailer != nil {
		log.Println("Encerrando serviço de e-mail...")
		_ = a.Mailer.Close()
	}
}

// Run inicia o servidor HTTP
func (a *App) Run() error {
	addr := fmt.Sprintf(":%s", a.Config.Port)
	log.Printf("🚀 Servidor Go rodando em http://localhost%s", addr)
	return http.ListenAndServe(addr, a.Router)
}
