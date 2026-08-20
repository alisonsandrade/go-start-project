// Package app onde estão centralizados todos os roteamentos e ciclo de vida do aplicativo
package app

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/alisonsandrade/go-start-project/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gorm.io/gorm"

	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/handler"
	"github.com/alisonsandrade/go-start-project/internal/repository"
	"github.com/alisonsandrade/go-start-project/internal/service"
)

type App struct {
	Config *config.Config
	DB     *gorm.DB
	Router *chi.Mux
}

func New() (*App, error) {
	// 1. Carrega configurações
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("falha ao carregar configs: %w", err)
	}

	// 2. Conecta ao Banco de Dados
	db, err := repository.NewDatabase(cfg)
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

	// 4. Dependency wiring
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	authService := service.NewAuthService(userRepo, tokenRepo, cfg)
	userService := service.NewUserService(userRepo)
	roleService := service.NewRoleService(roleRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	roleHandler := handler.NewRoleHandler(roleService)

	// 5. Route registration by domain
	r.Route("/api", func(api chi.Router) {
		api.Mount("/auth", authHandler.AuthRoutes(cfg))
		api.Mount("/users", userHandler.Routes(cfg, roleRepo))
		api.Mount("/roles", roleHandler.Routes(cfg, roleRepo))
	})

	// Documentação Swagger
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8000/swagger/doc.json"),
	))

	return &App{
		Config: cfg,
		DB:     db,
		Router: r,
	}, nil
}

// Run inicia o servidor HTTP
func (a *App) Run() error {
	addr := fmt.Sprintf(":%s", a.Config.Port)
	log.Printf("🚀 Servidor Go rodando em http://localhost%s", addr)
	return http.ListenAndServe(addr, a.Router)
}
