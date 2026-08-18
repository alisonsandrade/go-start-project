# Go Start Project 🚀

Boilerplate de API RESTful em **Go**, estruturado com Clean Architecture, autenticação JWT com refresh token rotativo, RBAC e pronto para produção. Clone, configure o `.env` e comece a construir o seu domínio.

> 🎯 **Filosofia:** modularidade pragmática sobre abstrações pesadas. Organização por camadas, código explícito e sem dependências desnecessárias.

---

## 📑 Índice

- [Tecnologias](#️-tecnologias-utilizadas)
- [Arquitetura](#️-arquitetura)
- [Estrutura do projeto](#️-estrutura-do-projeto)
- [Variáveis de ambiente](#️-variáveis-de-ambiente-env)
- [Como executar](#-como-executar)
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
| **Roteador** | Chi Router v5 |
| **ORM** | GORM |
| **Banco de Dados** | PostgreSQL 16 |
| **Autenticação** | JWT (`golang-jwt/jwt/v5`) + Refresh Token rotativo |
| **Autorização** | RBAC Middleware (ADMIN / USER) |
| **Segurança** | Bcrypt (`golang.org/x/crypto/bcrypt`) |
| **Documentação** | Swagger (`swaggo/swag`) |
| **Containerização** | Docker & Docker Compose |

---

## 🏛️ Arquitetura

O projeto segue uma separação clara em camadas, com injeção de dependências centralizada no *Composition Root* (`internal/app/app.go`):

```
handler  →  service  →  repository  →  banco de dados
```

- **handler** — recebe o HTTP, valida a entrada e traduz erros de negócio em status HTTP.
- **service** — regras de negócio (casos de uso). Isolado por domínio.
- **repository** — persistência (GORM). Exposto por interfaces.
- **domain** — entidades, DTOs e regras puras (inclui a validação dos DTOs).

Os domínios de **autenticação** e **gestão de usuários** são separados:

- `AuthService` / `AuthHandler` → registro, login, refresh e logout.
- `UserService` / `UserHandler` → perfil do usuário e administração.

---

## 🗂️ Estrutura do Projeto

```
.
├── cmd/
│   └── api/                 # Entrypoint principal (main.go)
├── internal/
│   ├── app/                 # Composition Root & ciclo de vida
│   ├── config/              # Variáveis de ambiente (.env)
│   ├── domain/              # Entidades, DTOs e validações
│   ├── handler/             # Controladores HTTP (auth e user)
│   ├── middleware/          # Autenticação JWT e RBAC
│   ├── repository/          # Persistência e conexão GORM
│   └── service/             # Casos de uso (auth e user)
├── migrations/              # Migrations SQL (up/down)
├── pkg/
│   └── token/               # Utilitários JWT (generate/parse)
├── docs/                    # Swagger gerado (não editar à mão)
├── docker-compose.yml       # PostgreSQL + API
├── Dockerfile               # Build multi-stage enxuto
├── LICENSE                  # Licença MIT
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
JWT_SECRET=super_secret_jwt_key_here
JWT_EXPIRATION_HOURS=24
```

> 🔒 **Segurança:** o `.env` está no `.gitignore` e **nunca** deve ser versionado. Em produção, use um `JWT_SECRET` longo (> 32 caracteres) e aleatório, injetado por um gerenciador de segredos.

---

## 🚀 Como Executar

### 1. Suba o PostgreSQL via Docker

```bash
docker-compose up -d postgres
```

### 2. Rode a aplicação

```bash
go run cmd/api/main.go
```

A API estará acessível em: **http://localhost:8000**

---

## 🗃️ Migrations

As migrations SQL ficam em `migrations/` (formato `up`/`down`). Elas criam a extensão `uuid-ossp`, a tabela `users` e a tabela `refresh_tokens`.

Aplique-as com a ferramenta [`golang-migrate`](https://github.com/golang-migrate/migrate):

```bash
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/gostartdb?sslmode=disable" up
```

---

## 📡 Endpoints da API

### 🔐 Autenticação (`/api/auth`)

| Método | Rota | Descrição | Acesso |
|---|---|---|---|
| `POST` | `/api/auth/register` | Cadastro de usuário (sempre role USER) | Público |
| `POST` | `/api/auth/login` | Autenticação e emissão de tokens | Público |
| `POST` | `/api/auth/refresh` | Rotaciona a sessão (novo par de tokens) | Público* |
| `POST` | `/api/auth/logout` | Invalida a sessão do usuário | JWT |

\* Requer um refresh token válido no corpo da requisição.

### 👤 Usuários (`/api/users`)

| Método | Rota | Descrição | Acesso |
|---|---|---|---|
| `GET` | `/api/users/me` | Perfil do usuário autenticado | JWT |
| `PUT` | `/api/users/me` | Atualiza o próprio perfil | JWT |
| `DELETE` | `/api/users/me` | Desativa a própria conta (soft delete) | JWT |
| `GET` | `/api/users` | Lista todos os usuários | ADMIN |

---

## 🔄 Fluxo de Autenticação

1. **Registro/Login** → o cliente recebe um `access_token` (curta duração) e um `refresh_token` (longa duração).
2. **Requisições autenticadas** → enviar o header `Authorization: Bearer <access_token>`.
3. **Expiração** → quando o `access_token` expira, use o `/api/auth/refresh` com o `refresh_token` para obter um novo par.
4. **Rotação** → a cada refresh, o token antigo é **revogado** e um novo é emitido (segurança contra reuso).
5. **Logout** → revoga **todos** os refresh tokens do usuário.

---

## 📖 Documentação Swagger

A documentação é gerada a partir das annotations nos handlers. Para regenerar após alterações:

```bash
swag init -g cmd/api/main.go
```

Com a aplicação rodando, acesse a UI interativa em:

**http://localhost:8000/swagger/index.html**

> ⚠️ Os arquivos em `docs/` são **gerados automaticamente** — não os edite à mão.

---

## 📄 Licença

Distribuído sob a licença **MIT**. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.
