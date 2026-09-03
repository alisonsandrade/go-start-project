# Go Start Project 🚀

Boilerplate de API RESTful de alta performance em **Go**, estruturado sob os princípios de **Clean Architecture** e **Domain-Driven Design (DDD)**. Conta com autenticação JWT, rotação de refresh tokens, controle de acesso baseado em papéis e permissões granulares (RBAC), Value Objects com validação de domínio nativa, rate limiting contra força bruta, mensageria interna assíncrona (Worker Pool) e auditoria de requisições mutativas.

> 🎯 **Filosofia:** Pragmático, robusto e livre de dependências externas desnecessárias. Concorrência nativa de Go para tarefas em background, arquitetura orientada a contextos de domínio e tipagem estrita com foco em produção.

---

## 📑 Índice

- [Tecnologias Utilizadas](#️-tecnologias-utilizadas)
- [Arquitetura & Diferenciais de Engenharia](#-arquitetura--diferenciais-de-engenharia)
- [Guia Rápido: Iniciando o Projeto do Zero](#-guia-rápido-iniciando-o-projeto-do-zero)
- [Estrutura do Diretório](#-estrutura-do-diretório)
- [Variáveis de Ambiente (.env)](#️-variáveis-de-ambiente-env)
- [Migrations & Banco de Dados](#-migrations--banco-de-dados)
- [Endpoints da API](#-endpoints-da-api)
- [Mecanismos de Segurança e Concorrência](#-mecanismos-de-segurança-e-concorrência)
- [Documentação Swagger](#-documentação-swagger)
- [Licença](#-licença)

---

## 🛠️ Tecnologias Utilizadas

| Camada | Tecnologia |
|---|---|
| **Linguagem** | Go 1.25+ |
| **Roteador HTTP** | Chi Router v5 (`go-chi/chi/v5`) |
| **ORM & Driver** | GORM com PostgreSQL (`gorm.io/gorm`, `gorm.io/driver/postgres`) |
| **Banco de Dados** | PostgreSQL 16 |
| **Autenticação** | JWT (`golang-jwt/jwt/v5`) + Refresh Tokens com invalidação em cascata |
| **Tráfego & Defesa** | Rate Limiting por conexão TCP (`go-chi/httprate`) imune a spoofing |
| **Concorrência** | Goroutines, Channels bufferizados e `sync.WaitGroup` |
| **Criptografia** | Bcrypt + SHA-256 + Tokens criptograficamente seguros (`crypto/rand`) |
| **Hot Reload** | Air (`.air.toml`) com Graceful Shutdown integrado |
| **Documentação** | Swagger (`swaggo/swag`, `swaggo/http-swagger/v2`) |
| **Containerização** | Docker & Docker Compose |

---

## 🏛️ Arquitetura & Diferenciais de Engenharia

O ecossistema é dividido em contextos isolados (`internal/auth`, `internal/users`, `internal/roles`, `internal/audit`):

```text
HTTP Request → Middleware (RateLimit, JWT, Audit) → Handler → Service → Repository → DB
                                                                ↓
                                                     Channel Buffer (100)
                                                                ↓
                                                    Worker Pool (3 Goroutines)

```

* **Propagação de `context.Context**`: Cancelamento instantâneo de queries e conexões I/O caso a chamada HTTP seja abortada pelo cliente.
* **Worker Pool Assíncrono (`pkg/mailer`)**: Fila de envio de e-mails interna com canal bufferizado (backpressure) e 3 workers em background. Dispensa brokers externos para envio de e-mails transacionais.
* **Graceful Shutdown**: Interceptação de sinais `SIGINT`/`SIGTERM`. Aguarda a finalização dos workers e encerra o socket HTTP sem quebrar pacotes ou travar portas no SO.
* **Value Objects com Validação Rica (`pkg/domain`)**: `Email` e `Password` encapsulam validação, normalização e hashing antes de tocar a camada de dados.
* **Auditoria Assíncrona (`internal/audit`)**: Middleware que intercepta ações mutativas (`POST`, `PUT`, `DELETE`), extrai o autor via Claims/Bearer Token e grava registros em background sem onerar a latência HTTP.
* **Prevenção de Account Enumeration**: Rotas públicas de recuperação de senha e cadastro respondem de forma indistinguível para e-mails inexistentes.

---

## 🚀 Guia Rápido: Iniciando o Projeto do Zero

Se você acabou de clonar este repositório em uma nova máquina, siga este roteiro:

### 1. Pré-requisitos

* **Go** (versão 1.25 ou superior)
* **Docker** e **Docker Compose**
* Ferramenta **golang-migrate** (opcional, caso não use Makefile)
* **Air** (para live-reload): `go install github.com/air-verse/air@latest`
* **Swag CLI**: `go install github.com/swaggo/swag/cmd/swag@latest`

### 2. Clonar e Instalar Dependências

```bash
git clone [https://github.com/seu-usuario/go-start-project.git](https://github.com/seu-usuario/go-start-project.git)
cd go-start-project

# Baixa as dependências e sincroniza o go.mod
go mod tidy

```

### 3. Configurar o Ambiente (.env)

Copie o exemplo de variáveis de ambiente:

```bash
cp .env.example .env

```

*(Edite o `.env` caso precise alterar portas ou credenciais de banco).*

### 4. Subir a Infraestrutura (Docker)

Inicie a instância do PostgreSQL:

```bash
docker compose up -d postgres

```

### 5. Executar as Migrations

Aplique todas as tabelas e índices no PostgreSQL:

```bash
make migrate-up
# Ou via CLI nativa do migrate:
# migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/gostartdb?sslmode=disable" up

```

### 6. Executar o Servidor

**Modo Desenvolvimento (com Live Reload do Air):**

```bash
air
# ou
make dev

```

**Modo Padrão (compilação direta):**

```bash
go run cmd/api/main.go
# ou
make run

```

A API estará acessível em: **http://localhost:8000**

Swagger interativo em: **http://localhost:8000/swagger/index.html**

---

## 🗂️ Estrutura do Diretório

```text
.
├── cmd/
│   └── api/
│       └── main.go                    # Entrypoint da aplicação
├── internal/
│   ├── app/
│   │   └── app.go                     # Composition Root, DI e Graceful Shutdown
│   ├── audit/                         # Módulo de Auditoria de Ações Mutativas
│   │   ├── audit.go                   # Entidade Log e Repositório
│   │   └── middleware.go              # Middleware de Auditoria Assíncrono
│   ├── auth/                          # Contexto de Autenticação & Sessões
│   │   ├── domain/                    # DTOs, Tokens e modelos de sessão
│   │   ├── errors.go                  # Sentinel errors de autenticação
│   │   ├── handler.go                 # Handlers (Login, Forgot, Reset, Change)
│   │   ├── middleware.go              # Middleware de autorização JWT
│   │   ├── repository.go              # Persistência de Refresh Tokens
│   │   └── service.go                 # Regras de negócio de credenciais
│   ├── config/
│   │   └── config.go                  # Parser de variáveis de ambiente
│   ├── platform/
│   │   ├── database/
│   │   │   └── db.go                  # Conexão GORM e pooling PostgreSQL
│   │   ├── ratelimit.go               # Limitador de tráfego seguro contra spoofing
│   │   └── response.go                # Utilitários de resposta JSON uniforme
│   ├── roles/                         # RBAC (Perfis e Permissões Granulares)
│   │   ├── domain/                    # Entidades RoleEntity e Permission
│   │   ├── errors.go                  # Erros de domínio para roles
│   │   ├── handler.go                 # Rotas administrativas de roles
│   │   ├── middleware.go              # Verificador de permissões (RequirePermission)
│   │   ├── permission_repository.go   # Consultas de permissões
│   │   ├── repository.go              # Associação many-to-many de papéis
│   │   └── service.go                 # Gestão de permissões e regras imutáveis
│   └── users/                         # Gestão de Usuários
│       ├── domain/                    # Entidade User com hooks de validação GORM
│       ├── handler.go                 # Rotas /me e administrativas
│       ├── repository.go              # Queries e Soft Delete
│       ├── seed.go                    # Seed idempotente do Administrador inicial
│       └── service.go                 # Regras de atualização e regras de usuário
├── migrations/                        # Arquivos SQL Up/Down ordenados
├── pkg/
│   ├── apiresponse/                   # Schemas uniformes de retorno da API
│   ├── domain/                        # Value Objects (Email, Password)
│   ├── mailer/                        # Worker Pool assíncrono para e-mails
│   └── token/                         # Emissor e validador de JWT Custom Claims
├── docs/                              # Arquivos gerados do Swagger
├── .air.toml                          # Configuração afinada para compilação local
├── Makefile                           # Comandos rápidos de compilação e migração
└── README.md

```

---

## ⚙️ Variáveis de Ambiente (.env)

```env
PORT=8000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gostartdb
DB_SSLMODE=disable

# JWT
JWT_SECRET=sua_chave_secreta_jwt_longa_e_aleatoria_com_mais_de_32_bytes
JWT_EXPIRATION_HOURS=24

# Seed Inicial (Criação automática do primeiro ADMIN)
ADMIN_DEFAULT_NAME=Super Admin
ADMIN_DEFAULT_EMAIL=admin@admin.com
ADMIN_DEFAULT_PASSWORD=Admin@123456

```

---

## 🗃️ Migrations & Banco de Dados

Tabelas versionadas sequencialmente em `migrations/`:

* `000001_create_users_table`: Tabela de usuários e extensão `uuid-ossp`/`pgcrypto`.
* `000002_create_refresh_tokens_table`: Sessões persistidas de refresh token.
* `000003_create_rbac_tables`: Tabelas `roles`, `permissions` e pivot `role_permissions`.
* `000007_create_audit_logs`: Tabela `audit_logs` com índices em `user_id` e `created_at`.

Comandos:

```bash
make migrate-up      # Roda todas as migrações para a versão mais recente
make migrate-down    # Reverte a última migração aplicada

```

---

## 📡 Endpoints da API

### 🔐 Autenticação (`/api/auth`)

*Endpoints públicos protegidos com Rate Limiting estrito.*

| Método | Rota | Descrição | Acesso |
| --- | --- | --- | --- |
| `POST` | `/api/auth/register` | Cria uma nova conta com perfil padrão | Público (Rate Limited) |
| `POST` | `/api/auth/login` | Emite tokens JWT e Refresh Token | Público (Rate Limited) |
| `POST` | `/api/auth/refresh` | Rotação de sessão: invalida o atual e emite novo par | Público |
| `POST` | `/api/auth/forgot-password` | Dispara token seguro temporário por e-mail | Público (Rate Limited) |
| `POST` | `/api/auth/reset-password` | Redefine senha com token e invalida sessões antigas | Público (Rate Limited) |
| `POST` | `/api/auth/logout` | Revoga todos os tokens ativos do usuário | JWT |
| `POST` | `/api/auth/change-password` | Altera senha autenticada (exige senha anterior) | JWT |

### 👤 Usuários (`/api/users`)

| Método | Rota | Descrição | Acesso |
| --- | --- | --- | --- |
| `GET` | `/api/users/me` | Dados cadastrais do próprio usuário autenticado | JWT |
| `PUT` | `/api/users/me` | Atualiza os dados do próprio usuário | JWT |
| `DELETE` | `/api/users/me` | Soft delete/desativação da própria conta | JWT |
| `GET` | `/api/users` | Lista todos os usuários | Permissão `user:list` |
| `POST` | `/api/users` | Cadastro administrativo de novos usuários | Permissão `user:create` |
| `GET` | `/api/users/{id}` | Busca perfil completo de usuário por UUID | Permissão `user:read` |
| `PUT` | `/api/users/{id}` | Edição de permissões/dados de outro usuário | Permissão `user:update` |
| `DELETE` | `/api/users/{id}` | Desativação forçada de usuário por UUID | Permissão `user:delete` |

### 🛡️ Papéis & Permissões (`/api/roles`)

| Método | Rota | Descrição | Acesso |
| --- | --- | --- | --- |
| `GET` | `/api/roles` | Lista todos os perfis disponíveis no sistema | Permissão `role:read` |
| `POST` | `/api/roles` | Criação de novo perfil customizado | Permissão `role:create` |
| `GET` | `/api/roles/{id}` | Detalhes de um perfil e permissões associadas | Permissão `role:read` |
| `PUT` | `/api/roles/{id}` | Altera nome/descrição (roles de sistema são imutáveis) | Permissão `role:update` |
| `DELETE` | `/api/roles/{id}` | Remove perfil (exceto papéis `is_system`) | Permissão `role:delete` |
| `PUT` | `/api/roles/{id}/permissions` | Substitui a lista de permissões associadas ao perfil | Permissão `role:assign-permissions` |

---

## 🔒 Mecanismos de Segurança e Concorrência

1. **Rate Limiting Imune a Spoofing**: O middleware extrai o endereço físico da conexão (`net.SplitHostPort(r.RemoteAddr)`). Isso impede que invasores utilizem cabeçalhos falsificados (`X-Forwarded-For` ou `X-Real-IP`) para burlar contadores ou orquestrar negação de serviço contra IPs de terceiros.
2. **Hashes Unidirecionais para Tokens de Recuperação**: O token temporário enviado ao e-mail do usuário não fica gravado no banco de dados. Armazena-se apenas o hash `SHA-256` da string, neutralizando o uso dos tokens em eventuais vazamentos de dumps de banco.
3. **Invalidação de Sessão em Cascata**: Ao executar um reset de senha, todos os *Refresh Tokens* emitidos para o `user_id` são expurgados do banco, deslogando instantaneamente qualquer invasor que esteja usando uma sessão anterior.
4. **Resiliência do Pipeline de Mensageria**: Caso o servidor precise ser reiniciado ou atualizado, o `sync.WaitGroup` garante que requisições de e-mail em curso sejam processadas até o fim antes da destruição dos canais de memória.

---

## 📖 Documentação Swagger

Annotations do Swagger são declaradas diretamente sobre os handlers HTTP. Para regenerar a documentação após modificar rotas ou contratos DTO:

```bash
make swag
# ou diretamente via swag CLI:
swag init -g cmd/api/main.go

```

Com o servidor em execução, acesse a interface interativa em:
👉 **http://localhost:8000/swagger/index.html**

---

## 📄 Licença

Distribuído sob a licença **MIT**. Consulte o arquivo `LICENSE` para mais detalhes.

```
