# Dockerfile - Multi-stage build optimizado

# ========================================
# Stage 1: Builder
# ========================================
FROM golang:1.22-alpine AS builder

# Instalar dependencias del sistema
RUN apk add --no-cache git ca-certificates tzdata

# Establecer working directory
WORKDIR /app

# Copiar go.mod y go.sum
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar binario
# CGO_ENABLED=0 para binario estático
# -ldflags="-w -s" para reducir tamaño
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/auth-service \
    cmd/server/main.go

# ========================================
# Stage 2: Runtime
# ========================================
FROM alpine:latest

# Instalar certificados CA para HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Crear usuario no-root
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

# Establecer working directory
WORKDIR /app

# Copiar binario desde builder
COPY --from=builder /app/bin/auth-service /app/auth-service

# Copiar configs
COPY --from=builder /app/configs /app/configs

# Cambiar ownership a appuser
RUN chown -R appuser:appgroup /app

# Cambiar a usuario no-root
USER appuser

# Exponer puerto
EXPOSE 4001

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:4001/health || exit 1

# Comando de inicio
CMD ["/app/auth-service"]