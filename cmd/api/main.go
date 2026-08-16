package main

import (
	"log"

	"github.com/alisonsandrade/go-start-project/internal/app"
)

func main() {
	// Cria o container da aplicação com todas as dependências resolvidas
	application, err := app.New()
	if err != nil {
		log.Fatalf("Erro ao inicializar a aplicação: %v", err)
	}

	// Sobe o servidor HTTP
	if err := application.Run(); err != nil {
		log.Fatalf("Erro na execução do servidor: %v", err)
	}
}
