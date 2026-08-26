# Go Start Project 🚀

Boilerplate de API RESTful em **Go**, estruturado com Clean Architecture e Domain-Driven Design (DDD), autenticação JWT com refresh token rotativo, RBAC modular, Value Objects com persistência nativa e propagação de `context.Context`.

> 🎯 **Filosofia:** modularidade pragmática sobre abstrações pesadas. Organização por contextos/módulos, código explícito, tipagem estrita com Value Objects e controle de ciclo de vida de requisições.

---

## 📑 Índice

- [Tecnologias](#️-tecnologias-utilizadas)
- [Arquitetura e Melhorias](#️-arquitetura-e-melhorias)
- [Estrutura do projeto](#️-estrutura-do-projeto)
- [Variáveis de ambiente](#️-variáveis-de-ambiente-env)
- [Como executar](#-como-executar)
- [Seed Inicial](#-seed-inicial-de-admin)
- [Migrations](#️-migrations)
- [Endpoints da API](#-endpoints-da-api)
- [Fluxo de autenticação](#-fluxo-de-autenticação)
- [Documentação Swagger](#-documentação-swagger)
- [Licença](#-licença)

---

## 🛠️ Tecnologias Utilizadas

| Camada | Tecnologia |
|---|---|
| **Linguagem** | Go 1.25+ |
| **Roteador & Middlewares** | Chi Router v5 (`go-chi/chi/v5` e `go-chi/cors`) |
| **ORM & Driver** | GORM com PostgreSQL (`gorm.io/gorm`, `gorm.io/driver/postgres`) |
| **Banco de Dados** | PostgreSQL 16 |
| **Autenticação** | JWT (`golang-jwt/jwt/v5`) + Refresh Token rotativo |
| **Autorização** | RBAC Modular (`auth`, `roles`, `users`) |
| **Criptografia & Tipos** | Bcrypt (`golang.org/x/crypto/bcrypt`) + UUID (`google/uuid`) |
| **Hot Reload** | Air (`.air.toml`) |
| **Documentação** | Swagger (`swaggo/swag`, `swaggo/http-swagger/v2`) |
| **Containerização** | Docker & Docker Compose |

---

## 🏛️ Arquitetura e Melhorias

O projeto adota a separação por **módulos de contexto** dentro de `internal/`, além de padrões sólidos da linguagem Go:


```

Handler (HTTP)  →  Service (Regras de Negócio)  →  Repository (GORM/SQL)

```

- **Propagação de `context.Context`**: Todos os métodos de repositories e services recebem `ctx context.Context`, garantindo cancelamento automático de queries de I/O em caso de disconnect/timeout do cliente HTTP.
- **Value Objects Centrais (`pkg/domain`)**: Tipos de domínio `Email` e `Password` com auto-validação, normalização, hashing e implementação nativa das interfaces `driver.Valuer`, `sql.Scanner` e `json.Marshaler`.
- **Factory de Entidade**: Construtores como `NewUser` asseguram que nenhuma struct nasça em estado inválido.
- **Seed Idempotente**: Rotina automática no bootstrap da aplicação para criação inicial do Administrador padrão sem falhas de duplicação.
- **Respostas JSON Padronizadas**: Módulo de resposta estruturada (`pkg/apiresponse`).

---

## 🗂️ Estrutura do Projeto

```text
.
├── cmd/
│   └── api/
│       └── main.go                    # Entrypoint da aplicação
├── internal/
│   ├── app/
│   │   └── app.go                     # Composition Root & ciclo de vida
│   ├── auth/                          # Contexto de Autenticação
│   │   ├── domain/                    # DTOs e Tokens
│   │   ├── errors.go                  # Erros de autenticação
│   │   ├── handler.go                 # Controladores HTTP de Auth
│   │   ├── middleware.go              # Middleware de Autenticação JWT
│   │   ├── repository.go              # Persistência de Refresh Tokens
│   │   └── service.go                 # Regras de Login/Refresh/Logout
│   ├── config/
│   │   └── config.go                  # Carregamento do .env
│   ├── platform/
│   │   ├── database/
│   │   │   └── db.go                  # Conexão e pooling GORM
│   │   └── response.go                # Helpers de resposta HTTP
│   ├── roles/                         # Contexto de RBAC / Permissões
│   │   ├── domain/                    # Entidades e DTOs de Roles
│   │   ├── errors.go                  # Erros de papéis
│   │   ├── handler.go                 # Controladores HTTP de Roles
│   │   ├── middleware.go              # Middleware de RBAC
│   │   ├── permission_repository.go   # Repositório de Permissões
│   │   ├── repository.go              # Repositório de Roles
│   │   └── service.go                 # Casos de uso de Roles
│   └── users/                         # Contexto de Gestão de Usuários
│       ├── domain/
│       │   ├── dto.go                 # DTOs de criação/atualização
│       │   └── user.go                # Entidade User e Factory
│       ├── handler.go                 # Controladores HTTP de Usuários
│       ├── repository.go              # Persistência de Usuários
│       ├── seed.go                    # Rotina de Seed do Admin
│       └── service.go                 # Casos de uso de Usuários
├── migrations/                        # Scripts SQL de migração (up/down)
├── pkg/
│   ├── apiresponse/                   # Padronização de mensagens e erros HTTP
│   ├── domain/                        # Value Objects compartilhados (Email, Password)
│   └── token/                         # Geração e validação de JWT
├── docs/                              # Swagger gerado automaticamente
├── .air.toml                          # Configuração de Live Reload com Air
├── Dockerfile                         # Build multi-stage da aplicação
├── Makefile                           # Automações de build, run, docs e migrations
├── LICENSE                            # Licença MIT
└── README.md

```

---

## ⚙️ Variáveis de Ambiente (.env)

Crie um arquivo `.env` na raiz do projeto:

```env
PORT=8000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gostartdb
DB_SSLMODE=disable

# JWT Settings
JWT_SECRET=super_secret_jwt_key_here
JWT_EXPIRATION_HOURS=24

# Seed Settings (Admin Padrão)
ADMIN_DEFAULT_NAME=Super Admin
ADMIN_DEFAULT_EMAIL=admin@admin.com
ADMIN_DEFAULT_PASSWORD=Admin@123456

```

> 🔒 **Segurança:** O arquivo `.env` está configurado no `.gitignore`. Em produção, utilize credenciais seguras e variáveis injetadas pelo ambiente de hospedagem.

---

## 🚀 Como Executar

### 1. Suba o Banco de Dados com Docker

```bash
docker-compose up -d postgres

```

### 2. Execute as Migrations

```bash
make migrate-up
# ou
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/gostartdb?sslmode=disable" up

```

### 3. Inicie a Aplicação

Para rodar com hot-reload (Air):

```bash
air

```

Ou diretamente com Go / Makefile:

```bash
make run
# ou
go run cmd/api/main.go

```

A API estará acessível em: **http://localhost:8000**

---

## 🌱 Seed Inicial de Admin

A aplicação conta com um inicializador em `internal/users/seed.go`. Ao iniciar a aplicação (`app.New()`):

1. O sistema verifica se o e-mail configurado em `ADMIN_DEFAULT_EMAIL` já existe.
2. Caso não exista, cria automaticamente a conta de administrador inicial com a role `admin`, utilizando a senha criptografada via Value Object `Password`.
3. Se o usuário já estiver cadastrado, a rotina não realiza nenhuma alteração.

---

## 🗃️ Migrations

As migrações SQL ficam na pasta `migrations/` no formato `up`/`down`:

* `000001_create_users_table` (tabela de usuários e extensão uuid)
* `000002_create_refresh_tokens_table` (sessões e refresh tokens)
* `000003_create_rbac_tables` (roles e permissões)

Comandos utilitários no `Makefile`:

```bash
make migrate-up      # Aplica todas as migrations pendentes
make migrate-down    # Reverte a última migration

```

---

## 📡 Endpoints da API

### 🔐 Autenticação (`/api/auth`)

| Método | Rota | Descrição | Acesso |
| --- | --- | --- | --- |
| `POST` | `/api/auth/register` | Cadastro de novo usuário (padrão role `user`) | Público |
| `POST` | `/api/auth/login` | Autenticação e emissão de tokens | Público |
| `POST` | `/api/auth/refresh` | Rotação de sessão (novo par de tokens) | Público* |
| `POST` | `/api/auth/logout` | Invalida a sessão e revoga tokens | JWT |

* Requer refresh token válido no corpo da requisição.

### 👤 Usuários (`/api/users`)

| Método | Rota | Descrição | Acesso |
| --- | --- | --- | --- |
| `GET` | `/api/users/me` | Retorna o perfil do usuário autenticado | JWT |
| `PUT` | `/api/users/me` | Atualiza o próprio perfil | JWT |
| `DELETE` | `/api/users/me` | Desativa a conta do usuário (soft delete) | JWT |
| `POST` | `/api/users` | Cria usuário com role customizada | ADMIN |
| `GET` | `/api/users` | Lista todos os usuários cadastrados | ADMIN |
| `PUT` | `/api/users/{id}` | Atualização de dados de usuário por ID | ADMIN |

### 🛡️ Papéis & Permissões (`/api/roles`)

| Método | Rota | Descrição | Acesso |
| --- | --- | --- | --- |
| `GET` | `/api/roles` | Lista todas as roles | ADMIN |
| `POST` | `/api/roles` | Criação de novo papel/role | ADMIN |
| `PUT` | `/api/roles/{id}` | Edição de papel existente | ADMIN |

---

## 🔄 Fluxo de Autenticação

1. **Login** $\rightarrow$ O cliente envia e-mail e senha e recebe um `access_token` (JWT) e um `refresh_token`.
2. **Requisições Autenticadas** $\rightarrow$ Enviar header `Authorization: Bearer <access_token>`.
3. **Expiração do Access Token** $\rightarrow$ Utilizar a rota `/api/auth/refresh` enviando o `refresh_token`.
4. **Rotação de Refresh Tokens** $\rightarrow$ A cada refresh bem-sucedido, o token anterior é revogado e um novo par é gerado.
5. **Logout** $\rightarrow$ Invalida os refresh tokens associados à sessão.

---

## 📖 Documentação Swagger

A documentação da API é gerada automaticamente pelo Swag. Para atualizar após alterar annotations nos handlers:

```bash
make swag
# ou
go run [github.com/swaggo/swag/cmd/swag@latest](https://github.com/swaggo/swag/cmd/swag@latest) init -g cmd/api/main.go

```

Com o servidor rodando, acesse a UI interativa:

**http://localhost:8000/swagger/index.html**

---

## 📄 Licença

Distribuído sob a licença **MIT**. Veja o arquivo [LICENSE](https://www.google.com/search?q=LICENSE) para mais detalhes.