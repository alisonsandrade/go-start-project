-- Ativa a extensão do Postgres para gerar UUIDs automaticamente
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Dropa a tabela legada criada pelo GORM (Marco Zero)
DROP TABLE IF EXISTS users;

-- Cria a tabela com a tipagem estrita
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'USER',
    phone VARCHAR(50),
    avatar_url VARCHAR(255),
    job_title VARCHAR(100),
    bio TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Cria um índice para otimizar as consultas do GORM (Soft Delete)
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
