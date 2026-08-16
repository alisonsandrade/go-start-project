# Stage 1: Build da aplicação
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Baixa as dependências em cache
COPY go.mod go.sum ./
RUN go mod download

# Copia o código fonte e compila o binário estático
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api

# Stage 2: Imagem final enxuta
FROM alpine:3.20

WORKDIR /app

# Instala certificados SSL e timezone
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/api .
COPY .env .

EXPOSE 8000

CMD ["./api"]
