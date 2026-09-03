// Package app onde estão centralizados todos os roteamentos e ciclo de vida do aplicativo
package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/alisonsandrade/go-start-project/docs"
	"github.com/alisonsandrade/go-start-project/pkg/mailer"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gorm.io/gorm"

	"github.com/alisonsandrade/go-start-project/internal/audit"
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

	// 4. Dependency wiring
	auditRepo := audit.NewRepository(db)
	userRepo := users.NewUserRepository(db)
	tokenRepo := auth.NewTokenRepository(db)
	roleRepo := roles.NewRoleRepository(db)

	// Registra o middleware de auditoria antes das rotas
	r.Use(audit.Middleware(auditRepo, cfg.JWTSecret))

	// Inicializa o serviço de e-mail
	emailService := mailer.NewWorkerMailer(3, 100)

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
// Run inicia o servidor HTTP com suporte a Graceful Shutdown
func (a *App) Run() error {
	addr := fmt.Sprintf(":%s", a.Config.Port)

	server := &http.Server{
		Addr:    addr,
		Handler: a.Router,
	}

	// Canal para interceptar sinais de encerramento do SO / Air
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Canal de erro para capturar falhas na inicialização do servidor
	serverErrChan := make(chan error, 1)

	go func() {
		log.Printf("🚀 Servidor Go rodando em http://localhost%s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	// Aguarda ou um erro no servidor ou o sinal de encerramento
	select {
	case err := <-serverErrChan:
		return fmt.Errorf("falha ao iniciar servidor HTTP: %w", err)

	case sig := <-shutdownChan:
		log.Printf("Sinal recebido (%s). Encerrando servidor e liberando porta...", sig)

		// Cria contexto com timeout de 5 segundos para liberar os sockets abertos
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Aviso: Forçando fechamento do servidor: %v", err)
			_ = server.Close()
		}
	}

	return nil
}
