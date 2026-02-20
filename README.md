# FarmaNexo Auth Service

Microservicio de autenticacion para la plataforma FarmaNexo. Maneja exclusivamente el ciclo de vida de autenticacion: registro, login, logout, refresh de tokens y verificacion JWT.

## Responsabilidades

Este servicio se encarga **unicamente** de:

- Registro de usuarios (crear cuenta con credenciales)
- Login (autenticar y generar tokens JWT)
- Logout (revocar tokens)
- Refresh token (renovar access token)
- Verificacion de tokens (middleware JWT)

La gestion de perfiles de usuario (datos personales, avatar, preferencias) es responsabilidad de un servicio separado.

## Stack Tecnologico

| Componente | Tecnologia |
|---|---|
| Lenguaje | Go 1.25 |
| Framework HTTP | Chi v5 |
| Base de datos | PostgreSQL (GORM) |
| Cache / Blacklist | Redis |
| Tokens | JWT (golang-jwt/v5) |
| Eventos | AWS SQS |
| Configuracion | Viper |
| Logging | Zap (structured) |
| Documentacion API | Swagger (swag) |

## Arquitectura

**Clean Architecture + CQRS** con patron Mediator personalizado.

```
internal/
  domain/           Entidades, interfaces de repositorios y servicios
  application/      Commands, handlers, validators, pre/post-processors
  infrastructure/   PostgreSQL, Redis, JWT, SQS (implementaciones concretas)
  presentation/     Router HTTP, controllers, DTOs, middlewares
  shared/           ApiResponse, constantes, errores de dominio
pkg/
  mediator/         Mediator generico con pipeline CQRS
  config/           Carga de configuracion por ambiente
```

### Flujo de un request

```
HTTP Request
  -> Chi Router
    -> [AuthMiddleware si es protegido]
    -> Controller
      -> Mediator.Send(Command)
        -> Validator
        -> PreProcessor
        -> Handler
        -> PostProcessor
      <- ApiResponse[T]
    <- JSON Response
```

## API Endpoints

| Endpoint | Metodo | Auth | Descripcion | Rate Limit |
|---|---|---|---|---|
| `POST /api/v1/auth/register` | POST | Publico | Registrar nuevo usuario | - |
| `POST /api/v1/auth/login` | POST | Publico | Iniciar sesion | 5 req / 15 min (por email) |
| `POST /api/v1/auth/refresh` | POST | Publico | Renovar tokens | 10 req / 1 hora (por usuario) |
| `POST /api/v1/auth/logout` | POST | Protegido | Cerrar sesion | - |
| `GET /health` | GET | Publico | Health check | - |

### Formato de respuesta

Todos los endpoints retornan `ApiResponse[T]`:

```json
{
  "meta": {
    "mensajes": [
      {
        "codigo": "AUTH_002",
        "mensaje": "Inicio de sesion exitoso",
        "tipo": "exito"
      }
    ],
    "idTransaccion": "550e8400-e29b-41d4-a716-446655440000",
    "resultado": true,
    "timestamp": "2026-02-20T10:30:00Z"
  },
  "datos": {
    "access_token": "eyJhbGciOi...",
    "refresh_token": "eyJhbGciOi...",
    "token_type": "Bearer",
    "expires_in": 3600
  }
}
```

### Registro

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
  "datos": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "usuario@ejemplo.com",
    "message": "Usuario registrado exitosamente. Por favor inicia sesion.",
    "created_at": "2026-02-20T10:30:00Z"
  }
}
```

### Login

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
  "datos": {
    "access_token": "eyJhbGciOi...",
    "refresh_token": "eyJhbGciOi...",
    "token_type": "Bearer",
    "expires_in": 3600
  }
}
```

### Refresh Token

```bash
curl -X POST http://localhost:4001/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOi..."
  }'
```

### Logout

```bash
curl -X POST http://localhost:4001/api/v1/auth/logout \
  -H "Authorization: Bearer eyJhbGciOi..." \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOi..."
  }'
```

## Rate Limiting

El servicio implementa rate limiting con Redis usando el patron fixed-window counter:

| Endpoint | Key | Limite | Ventana |
|---|---|---|---|
| Login | `ratelimit:login:{email}` | 5 requests | 15 minutos |
| Refresh | `ratelimit:refresh:{user_id}` | 10 requests | 1 hora |

**Politica fail-open**: si Redis no esta disponible, el request pasa con un warning en logs.

Headers de rate limit en refresh:
- `X-RateLimit-Limit`: limite maximo
- `X-RateLimit-Remaining`: requests restantes
- `X-RateLimit-Reset`: timestamp Unix de reset

## Seguridad

### JWT

- **Access Token**: contiene `sub` (userID) + `role`. Duracion corta (configurable por ambiente).
- **Refresh Token**: contiene `sub` (userID) + `jti` (token ID). Duracion larga. Se almacena hasheado (SHA-256) en BD.
- **Rotacion de tokens**: al hacer refresh, el token anterior se revoca y se genera uno nuevo.
- **Blacklist**: los access tokens revocados (logout) se almacenan en Redis con TTL automatico.

### Passwords

- Hashing con bcrypt (cost default)
- Validacion minima: 8 caracteres

## Configuracion

### Ambientes

| Ambiente | Config | Descripcion |
|---|---|---|
| `local` | `configs/config.local.yaml` | Maquina del desarrollador (gitignored) |
| `development` | `configs/config.development.yaml` | Servidor de desarrollo |
| `qa` | `configs/config.qa.yaml` | Quality Assurance |
| `uat` | `configs/config.uat.yaml` | User Acceptance Testing |
| `production` | `configs/config.production.yaml` | Produccion |

Seleccion via variable de entorno `ENV` (default: `local`).

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
| `SQS_AUTH_EVENTS_QUEUE_URL` | URL de la cola SQS para eventos |

## Desarrollo

### Pre-requisitos

- Go 1.25+
- PostgreSQL 15+
- Redis 7+
- Docker y Docker Compose (para LocalStack)

### Setup local

1. Copiar configuracion local:
```bash
cp configs/config.development.yaml configs/config.local.yaml
# Editar config.local.yaml con valores locales
```

2. Aplicar migraciones:
```bash
make migrate-up
```

3. Ejecutar:
```bash
make dev
```

4. Swagger UI disponible en: http://localhost:4001/swagger/index.html

### Migraciones

```bash
make migrate-create NAME=add_new_table   # Crear migracion
make migrate-up                           # Aplicar pendientes
make migrate-down                         # Revertir ultima
```

Las migraciones estan en `migrations/` con formato golang-migrate. Schema: `auth`.

### Comandos utiles

```bash
make build           # Compilar binario
make test            # Ejecutar tests
make test-coverage   # Tests con reporte de cobertura
make lint            # Linter (golangci-lint)
make format          # Formatear codigo
make gen-mocks       # Generar mocks para tests
```

## Eventos

El servicio publica eventos de autenticacion via AWS SQS:

| Evento | Descripcion |
|---|---|
| `USER_REGISTERED` | Usuario registrado exitosamente |
| `USER_LOGIN` | Login exitoso |
| `USER_LOGOUT` | Logout exitoso |

Los eventos son fire-and-forget (best-effort). Si SQS falla, el flujo principal no se ve afectado.

## Estructura de directorios

```
auth-service/
  cmd/server/                    Punto de entrada (main.go)
  configs/                       Archivos YAML por ambiente
  internal/
    application/
      commands/                  RegisterUser, Login, RefreshToken, Logout
      handlers/                  Handler por cada command
      validators/                Validadores de commands
      preprocessors/             Sanitizacion de input
      postprocessors/            Auditoria de operaciones
    domain/
      entities/                  User, RefreshToken
      events/                    Eventos de autenticacion
      repositories/              Interfaces de repositorios
      services/                  Interfaces: TokenBlacklist, RateLimiter, EventPublisher
    infrastructure/
      cache/                     Redis: blacklist, rate limiter
      messaging/                 SQS event publisher
      persistence/postgres/      Repositorios GORM
      security/                  JWT service
    presentation/
      dto/requests/              RegisterRequest, LoginRequest, etc.
      dto/responses/             LoginResponse, RegisterResponse, EmptyResponse
      http/controllers/          AuthController
      http/middlewares/           AuthMiddleware, RateLimitMiddleware, CorrelationID
      http/routes/               Configuracion de rutas Chi
    shared/
      common/                    ApiResponse, response factories
      constants/                 Codigos HTTP, message codes
  migrations/                    SQL migrations (golang-migrate)
  pkg/
    config/                      Carga de configuracion (Viper)
    mediator/                    Mediator CQRS generico
  docs/                          Swagger generado
```
