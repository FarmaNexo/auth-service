# Auth Service

Microservicio de autenticacion para la plataforma FarmaNexo. Maneja exclusivamente el ciclo de vida de autenticacion: registro, login, logout, refresh de tokens y verificacion JWT.

**NO** gestiona perfiles de usuario, avatares ni preferencias. Esa responsabilidad es del User Service.

## Inicio Rapido

### Prerequisitos
- Go 1.25+
- PostgreSQL 16
- Redis 7
- LocalStack (desarrollo local)
- Docker & Docker Compose

### Instalacion
```bash
# Clonar repositorio
git clone <url>
cd services/auth-service

# Instalar dependencias
go mod download

# Configurar ambiente local
cp configs/config.development.yaml configs/config.local.yaml
# Editar configs/config.local.yaml con tus credenciales

# Crear base de datos
docker exec -it farmanexo-postgres psql -U admin -c "CREATE DATABASE auth_db;"

# Ejecutar migraciones
make migrate-up

# Ejecutar servicio
make dev
```

Swagger UI disponible en: http://localhost:4001/swagger/index.html

## Endpoints

### Publicos

**POST /api/v1/auth/register** - Registro de usuario
```bash
curl -X POST http://localhost:4001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@ejemplo.com",
    "password": "MiPassword123",
    "full_name": "Juan Perez",
    "phone": "+51999888777"
  }'
```

Respuesta (201):
```json
{
  "meta": {
    "mensajes": [{ "codigo": "AUTH_001", "mensaje": "Usuario registrado exitosamente", "tipo": "exito" }],
    "idTransaccion": "550e8400-e29b-41d4-a716-446655440000",
    "resultado": true,
    "timestamp": "20260222 103000"
  },
  "datos": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "usuario@ejemplo.com",
    "message": "Usuario registrado exitosamente. Por favor inicia sesion.",
    "created_at": "2026-02-22T10:30:00Z"
  }
}
```

**POST /api/v1/auth/login** - Iniciar sesion
```bash
curl -X POST http://localhost:4001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@ejemplo.com",
    "password": "MiPassword123"
  }'
```

Respuesta (200):
```json
{
  "meta": {
    "mensajes": [{ "codigo": "AUTH_002", "mensaje": "Inicio de sesion exitoso", "tipo": "exito" }],
    "idTransaccion": "...",
    "resultado": true,
    "timestamp": "20260222 103000"
  },
  "datos": {
    "access_token": "eyJhbGciOi...",
    "refresh_token": "eyJhbGciOi...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

- Rate limit: 5 requests / 15 minutos por email
- Error 429 si se excede

**POST /api/v1/auth/refresh** - Renovar tokens
```bash
curl -X POST http://localhost:4001/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOi..."
  }'
```

- Rate limit: 10 requests / 1 hora por usuario
- Headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`
- Rotacion automatica: el refresh token anterior se revoca

### Protegidos (requieren autenticacion)

**POST /api/v1/auth/logout** - Cerrar sesion
```bash
curl -X POST http://localhost:4001/api/v1/auth/logout \
  -H "Authorization: Bearer eyJhbGciOi..." \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOi..."
  }'
```

- El access token se agrega a la blacklist en Redis
- El refresh token se marca como revocado en BD

### Health Check

**GET /health** - Estado del servicio
```bash
curl http://localhost:4001/health
```

## Arquitectura

- **Puerto:** 4001
- **Base de datos:** `auth_db`
- **Schema:** `auth`
- **Patron:** Clean Architecture + CQRS + MediatR

### Capas

```
internal/
  domain/           Entidades (User, RefreshToken), interfaces de repositorios y servicios
  application/      Commands, Handlers, Validators, Pre/Post processors
  infrastructure/   PostgreSQL (GORM), Redis, JWT, SQS (implementaciones)
  presentation/     Controllers, Middlewares, Routes, DTOs
  shared/           ApiResponse[T], Constants, Errors
pkg/
  mediator/         Mediator CQRS generico con pipeline
  config/           Carga de configuracion por ambiente (Viper)
```

### Flujo de un request

```
HTTP Request
  -> Chi Router
    -> [Middlewares globales: RequestID, RealIP, Logger, Recoverer, CORS, CorrelationID]
    -> [AuthMiddleware si es protegido]
    -> Controller
      -> Mediator.Send(Command)
        -> Validator
        -> PreProcessor (SanitizeInput)
        -> Handler
        -> PostProcessor (LogAudit)
      <- ApiResponse[T]
    <- JSON Response
```

## Configuracion

### Variables de Entorno

```yaml
# configs/config.local.yaml
environment: local

server:
  host: 0.0.0.0
  port: 4001
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 60s

database:
  host: localhost
  port: 5432
  user: admin
  password: admin
  dbname: auth_db
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 5m

jwt:
  secret: "dev-super-secret-key-change-in-production-min-32-chars"
  access_token_duration: 15m
  refresh_token_duration: 168h
  issuer: "farmanexo-auth-service"

redis:
  host: localhost
  port: 6379
  password: farmanexo2026
  db: 0
  max_retries: 3
  pool_size: 10

aws:
  region: us-east-1
  endpoint: "http://localhost:4566"

sqs:
  auth_events_queue_url: "http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/farmanexo-auth-events"

log:
  level: debug
  encoding: console
```

### Ambientes

| Ambiente | Config | Seleccion |
|---|---|---|
| `local` | `configs/config.local.yaml` | Default (sin ENV) |
| `development` | `configs/config.development.yaml` | `ENV=development` |
| `qa` | `configs/config.qa.yaml` | `ENV=qa` |
| `uat` | `configs/config.uat.yaml` | `ENV=uat` |
| `production` | `configs/config.production.yaml` | `ENV=production` |

### Variables de entorno requeridas (ambientes desplegados)

| Variable | Descripcion |
|---|---|
| `DB_HOST` | Host de PostgreSQL |
| `DB_PORT` | Puerto de PostgreSQL |
| `DB_USER` | Usuario de PostgreSQL |
| `DB_PASSWORD` | Password de PostgreSQL |
| `DB_NAME` | Nombre de base de datos |
| `JWT_SECRET` | Secret para firmar JWT (min 32 chars) |
| `REDIS_HOST` | Host de Redis |
| `REDIS_PORT` | Puerto de Redis |
| `REDIS_PASSWORD` | Password de Redis |
| `AWS_REGION` | Region de AWS |
| `SQS_AUTH_EVENTS_QUEUE_URL` | URL de la cola SQS |

## Infraestructura

### PostgreSQL
- **Database:** `auth_db`
- **Schema:** `auth`
- **Tablas:** `users`, `refresh_tokens`

### Redis
- **Uso:** Blacklist de tokens (logout) + Rate limiting (login, refresh)
- **Keys:**
  - `blacklist:token:{jti}` - TTL igual a expiracion del access token
  - `ratelimit:login:{email}` - 5 req / 15 min
  - `ratelimit:refresh:{user_id}` - 10 req / 1 hora
- **Politica fail-open:** si Redis no esta disponible, el request pasa con warning en logs

### SQS
- **Publica en:** `farmanexo-auth-events`
- **Eventos:** `USER_REGISTERED`, `USER_LOGIN`, `USER_LOGOUT`
- **Patron:** fire-and-forget (best-effort, no bloquea el flujo principal)

### S3
- No utilizado por este servicio

## Seguridad

### JWT
- **Access Token:** contiene `sub` (userID) + `role`. Duracion corta (15min local, 1h dev, 15min prod)
- **Refresh Token:** contiene `sub` + `jti`. Duracion larga (7 dias). Hash SHA-256 almacenado en BD
- **Rotacion:** al hacer refresh, el token anterior se revoca
- **Blacklist:** tokens revocados (logout) se almacenan en Redis con TTL automatico

### Passwords
- Hashing con bcrypt (cost default)
- Validacion minima: 8 caracteres

### Rate Limiting

| Endpoint | Key | Limite | Ventana |
|---|---|---|---|
| Login | `ratelimit:login:{email}` | 5 requests | 15 minutos |
| Refresh | `ratelimit:refresh:{user_id}` | 10 requests | 1 hora |

## Testing
```bash
# Unit tests
make test

# Tests con coverage
make test-coverage

# Generar mocks
make gen-mocks
```

## Comandos Utiles
```bash
# Desarrollo
make dev              # Ejecutar en modo desarrollo
make build            # Compilar binario a bin/auth-service
make swagger          # Generar documentacion Swagger

# Base de datos
make migrate-up       # Aplicar migraciones pendientes
make migrate-down     # Revertir ultima migracion
make migrate-create NAME=nombre  # Crear nueva migracion

# Calidad
make lint             # Ejecutar golangci-lint
make format           # Formatear codigo con goimports

# Docker
make docker-build     # Construir imagen Docker
make docker-run       # Ejecutar container
```

## Dependencias

### Principales
- `github.com/go-chi/chi/v5` - HTTP router
- `gorm.io/gorm` - ORM
- `github.com/redis/go-redis/v9` - Redis client
- `github.com/aws/aws-sdk-go-v2` - AWS SDK (SQS)
- `github.com/golang-jwt/jwt/v5` - JWT
- `golang.org/x/crypto` - Bcrypt
- `go.uber.org/zap` - Structured logging
- `github.com/spf13/viper` - Configuracion
- `github.com/swaggo/swag` - Swagger

### Completas
Ver `go.mod`

## Eventos

| Evento | Trigger | Cola |
|---|---|---|
| `USER_REGISTERED` | Registro exitoso | `farmanexo-auth-events` |
| `USER_LOGIN` | Login exitoso | `farmanexo-auth-events` |
| `USER_LOGOUT` | Logout exitoso | `farmanexo-auth-events` |

Formato:
```json
{
  "event_type": "USER_REGISTERED",
  "user_id": "uuid",
  "email": "user@example.com",
  "timestamp": "2026-02-22T12:00:00Z",
  "metadata": {
    "source": "auth-service",
    "version": "1.0"
  }
}
```

## Documentacion Adicional

- [CLAUDE.md](./CLAUDE.md) - Contexto para Claude AI
- [INFRASTRUCTURE.md](./INFRASTRUCTURE.md) - Detalle de infraestructura
- [Swagger UI](http://localhost:4001/swagger/index.html) - API docs interactiva

## Estructura de Directorios

```
auth-service/
  cmd/server/main.go                    Punto de entrada con DI
  configs/                              YAML por ambiente (5 archivos)
  migrations/                           SQL (golang-migrate, schema: auth)
  internal/
    application/
      commands/                         RegisterUser, Login, RefreshToken, Logout
      handlers/                         Handler por cada command
      validators/                       RegisterUser, Login
      preprocessors/                    SanitizeInput
      postprocessors/                   LogAudit
    domain/
      entities/                         User, RefreshToken
      events/                           Eventos de autenticacion
      repositories/                     UserRepository, TokenRepository (interfaces)
      services/                         TokenBlacklist, RateLimiter, EventPublisher (interfaces)
    infrastructure/
      cache/                            Redis: blacklist, rate limiter
      messaging/                        SQS event publisher
      persistence/postgres/             Repositorios GORM
      security/                         JWT service
    presentation/
      dto/requests/                     RegisterRequest, LoginRequest, etc.
      dto/responses/                    LoginResponse, RegisterResponse, EmptyResponse
      http/controllers/                 AuthController
      http/middlewares/                 AuthMiddleware, RateLimitMiddleware, CorrelationID
      http/routes/                      Configuracion de rutas Chi
    shared/
      common/                           ApiResponse[T], response factories
      constants/                        Codigos HTTP, message codes
  pkg/
    config/                             Carga de configuracion (Viper)
    mediator/                           Mediator CQRS generico
  docs/                                 Swagger generado
```
