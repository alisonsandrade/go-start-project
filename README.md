# Go Start Project 🚀

Boilerplate de API RESTful em Go estruturado com Clean Architecture, alta performance e pronto para produção.

---

## 🛠️ Tecnologias Utilizadas

* Linguagem: Go (1.23+)
* Roteador: Chi Router v5
* ORM: GORM
* Banco de Dados: PostgreSQL 16
* Autenticação & Autorização: JWT (golang-jwt/jwt/v5) + RBAC Middleware
* Segurança: Bcrypt (golang.org/x/crypto/bcrypt)
* Containerização: Docker & Docker Compose

---

## 🏛️ Estrutura do Projeto

.
├── cmd/
│   └── api/                 # Entrypoint principal (main.go)
├── internal/
│   ├── app/                 # Application Container & Bootstrap
│   ├── config/              # Gerenciamento de variáveis de ambiente (.env)
│   ├── domain/              # Entidades, DTOs e regras de domínio puras
│   ├── handler/             # Controladores HTTP (request/response)
│   ├── middleware/          # Middlewares de Autenticação JWT e RBAC
│   ├── repository/          # Camada de persistência e conexão GORM
│   └── service/             # Casos de uso e regras de negócio
├── pkg/
│   └── token/               # Utilitários reaproveitáveis (JWT generator/parser)
├── docker-compose.yml       # Configuração do PostgreSQL
├── Dockerfile               # Build multi-stage enxuto
├── insomnia_collection.json # Coleção pronta de testes do Insomnia
├── LICENSE                  # Licença MIT
└── README.md

---

## ⚙️ Variáveis de Ambiente (.env)

Crie um arquivo `.env` na raiz do projeto com a seguinte configuração:

PORT=8000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gostartdb
DB_SSLMODE=disable
JWT_SECRET=super_secret_jwt_key_here
JWT_EXPIRATION_HOURS=24

---

## 🚀 Como Executar

### 1. Subir o PostgreSQL via Docker
docker-compose up -d postgres

### 2. Rodar a aplicação em Go
go run cmd/api/main.go

A API estará acessível em: http://localhost:8000

---

## 📡 Endpoints da API

| Método | Rota | Descrição | Acesso |
| :--- | :--- | :--- | :--- |
| POST | /api/users | Cadastro de novo usuário | Público |
| POST | /api/users/login | Autenticação e emissão de JWT | Público |
| GET | /api/users/me | Obter dados do perfil autenticado | Logado (JWT) |
| GET | /api/users | Listagem geral de usuários | Apenas ADMIN |

---

## 🧪 Testes via Insomnia

Importe o arquivo `insomnia_collection.json` diretamente no seu Insomnia para ter todas as rotas e ambientes pré-configurados.

---

## 📄 Licença

Distribuído sob a licença MIT. Veja o arquivo LICENSE para mais detalhes.
