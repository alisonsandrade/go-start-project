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

	// 4. Instanciação e Fiação dos Módulos
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg)
	userHandler := handler.NewUserHandler(userService)

	// 5. Registro de Rotas por Domínio
	r.Route("/api", func(api chi.Router) {
		api.Mount("/users", userHandler.Routes(cfg))
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
