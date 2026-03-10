# INFRAESTRUCTURA - Auth Service

## Resumen

Este documento describe la infraestructura especifica utilizada por Auth Service (puerto 4001).

---

## SERVICIOS REQUERIDOS

### PostgreSQL
- **Host:** localhost:5432 (local) / RDS endpoint (cloud)
- **Database:** `auth_db`
- **User:** admin (local) / `${DB_USER}` (cloud)
- **Password:** admin (local) / `${DB_PASSWORD}` (cloud)
- **Schema:** `auth`
- **Extensiones:** `uuid-ossp`
- **SSL:** disable (local) / require (produccion)

### Redis
- **Host:** localhost:6379 (local) / ElastiCache (cloud)
- **Password:** farmanexo2026 (local) / `${REDIS_PASSWORD}` (cloud)
- **DB:** 0 (compartido)
- **Pool size:** 10 (local) / 50 (produccion)
- **Max retries:** 3
- **Uso en este servicio:**
  - Blacklist de access tokens revocados (logout)
  - Rate limiting de login (por email)
  - Rate limiting de refresh (por user_id)

### LocalStack (Local) / AWS (Cloud)
- **Endpoint:** http://localhost:4566 (local)
- **Region:** us-east-1
- **Credenciales (local):** test/test (fake)

---

## RECURSOS AWS UTILIZADOS

### S3 Buckets

Auth Service **no utiliza S3**.

### SQS Queues

**Cola que PUBLICA:**

**Cola:** `farmanexo-auth-events`
- **URL (local):** `http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/farmanexo-auth-events`
- **URL (cloud):** `${SQS_AUTH_EVENTS_QUEUE_URL}`

**Eventos que genera:**

1. **USER_REGISTERED** - Cuando un usuario se registra exitosamente
```json
{
  "event_type": "USER_REGISTERED",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "timestamp": "2026-02-22T12:00:00Z",
  "metadata": {
    "source": "auth-service",
    "version": "1.0"
  }
}
```

2. **USER_LOGIN** - Cuando un usuario inicia sesion exitosamente
```json
{
  "event_type": "USER_LOGIN",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "timestamp": "2026-02-22T12:00:00Z",
  "metadata": {}
}
```

3. **USER_LOGOUT** - Cuando un usuario cierra sesion
```json
{
  "event_type": "USER_LOGOUT",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2026-02-22T12:00:00Z",
  "metadata": {}
}
```

**Patron de publicacion:** Fire-and-forget en goroutines. Si SQS falla, el flujo principal no se bloquea. Error se registra con `logger.Warn()`.

**Auth Service no consume eventos de ninguna cola.**

---

## ESQUEMA DE BASE DE DATOS

### Tabla: `auth.users`

```sql
CREATE TABLE auth.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    phone VARCHAR(50),
    is_active BOOLEAN DEFAULT true NOT NULL,
    is_verified BOOLEAN DEFAULT false NOT NULL,
    email_verified_at TIMESTAMPTZ,
    role VARCHAR(50) DEFAULT 'user' NOT NULL,
    last_login_at TIMESTAMPTZ,
    login_count INTEGER DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_users_email ON auth.users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON auth.users(deleted_at);
CREATE INDEX idx_users_role ON auth.users(role);
CREATE INDEX idx_users_is_active ON auth.users(is_active);
```

**Proposito:** Almacena credenciales y estado de autenticacion de usuarios.

**Roles validos:** `user`, `pharmacy_owner`, `admin`

**Soft delete:** Campo `deleted_at` (NULL = activo)

### Tabla: `auth.refresh_tokens`

```sql
CREATE TABLE auth.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    is_revoked BOOLEAN DEFAULT false NOT NULL,
    revoked_at TIMESTAMPTZ,
    ip_address VARCHAR(50),
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT fk_refresh_tokens_user
        FOREIGN KEY (user_id)
        REFERENCES auth.users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_refresh_tokens_user_id ON auth.refresh_tokens(user_id);
CREATE UNIQUE INDEX idx_refresh_tokens_token_hash ON auth.refresh_tokens(token_hash) WHERE is_revoked = false;
CREATE INDEX idx_refresh_tokens_expires_at ON auth.refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_is_revoked ON auth.refresh_tokens(is_revoked) WHERE is_revoked = false;
```

**Proposito:** Almacena hashes SHA-256 de refresh tokens. Nunca se guarda el token en texto plano.

**Relacion:** `user_id` -> `auth.users(id)` (CASCADE on delete)

---

## CONFIGURACION POR AMBIENTE

### Local (config.local.yaml)
```yaml
environment: local
server:
  port: 4001
  read_timeout: 30s
  write_timeout: 30s
database:
  host: localhost
  port: 5432
  user: admin
  password: admin
  dbname: auth_db
  sslmode: disable
  max_open_conns: 25
jwt:
  secret: "dev-super-secret-key-change-in-production-min-32-chars"
  access_token_duration: 15m
  refresh_token_duration: 168h  # 7 dias
redis:
  host: localhost
  port: 6379
  password: farmanexo2026
  pool_size: 10
aws:
  endpoint: "http://localhost:4566"
log:
  level: debug
  encoding: console
```

### Development (config.development.yaml)
```yaml
environment: development
database:
  host: ${DB_HOST}
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  dbname: auth_db_dev
  sslmode: disable
jwt:
  secret: ${JWT_SECRET}
  access_token_duration: 1h
aws:
  endpoint: ""  # AWS real
log:
  level: debug
  encoding: console
```

### Production (config.production.yaml)
```yaml
environment: production
server:
  read_timeout: 15s
  write_timeout: 15s
database:
  sslmode: require
  max_open_conns: 100
  max_idle_conns: 50
jwt:
  access_token_duration: 15m
  refresh_token_duration: 72h  # 3 dias
redis:
  pool_size: 50
log:
  level: info
  encoding: json
```

---

## SECRETS Y CREDENCIALES

### Secrets Manager (Produccion)
- `farmanexo/auth/jwt-secret` - JWT signing key
- `farmanexo/database/password` - Database password

### Variables de Entorno (Ambientes desplegados)

| Variable | Descripcion |
|---|---|
| `ENV` | local, development, qa, uat, production |
| `DB_HOST` | Host PostgreSQL |
| `DB_PORT` | Puerto PostgreSQL |
| `DB_USER` | Usuario PostgreSQL |
| `DB_PASSWORD` | Password PostgreSQL |
| `DB_NAME` | Nombre de base de datos |
| `JWT_SECRET` | JWT signing secret (min 32 chars) |
| `REDIS_HOST` | Host Redis |
| `REDIS_PORT` | Puerto Redis |
| `REDIS_PASSWORD` | Password Redis |
| `AWS_REGION` | Region AWS |
| `SQS_AUTH_EVENTS_QUEUE_URL` | URL cola SQS |

---

## CACHE REDIS - PATRONES

### Keys Utilizadas

| Pattern | TTL | Uso |
|---------|-----|-----|
| `blacklist:token:{jti}` | Igual a expiracion del access token | Tokens revocados por logout |
| `ratelimit:login:{email}` | 15 minutos | Contador de intentos de login por email |
| `ratelimit:refresh:{user_id}` | 1 hora | Contador de refresh por usuario |

### Politica Fail-Open

Si Redis no esta disponible:
- **Blacklist:** El request pasa (warning en logs). Riesgo: token revocado podria usarse hasta expirar
- **Rate limiting:** El request pasa (warning en logs). Riesgo: no hay limitacion temporal

### Invalidacion de Cache
- **Blacklist:** Se auto-invalida por TTL (no se borra manualmente)
- **Rate limit:** Se auto-invalida por TTL (ventana fija)

---

## EVENTOS SQS

### Flujo de Eventos

```
[Auth Service] --USER_REGISTERED--> [farmanexo-auth-events] --> [User Service] (crea perfil inicial)
[Auth Service] --USER_LOGIN------> [farmanexo-auth-events] --> [Consumidores futuros]
[Auth Service] --USER_LOGOUT-----> [farmanexo-auth-events] --> [Consumidores futuros]
```

### Procesamiento
- Los eventos se publican de forma asincrona en goroutines
- No hay garantia de entrega (best-effort)
- Si SQS esta caido, el evento se pierde pero el flujo principal funciona
- Los errores de publicacion se loguean con `logger.Warn()`

---

## DESPLIEGUE

### Checklist Pre-Deploy
- [ ] Migraciones aplicadas
- [ ] Variables de entorno configuradas
- [ ] JWT_SECRET en AWS Secrets Manager (min 32 chars)
- [ ] SQS queue `farmanexo-auth-events` creada
- [ ] Redis accesible
- [ ] PostgreSQL accesible con database `auth_db`

### Comandos de Deploy
```bash
# Build
make build

# Migraciones
make migrate-up ENV=production

# Docker
make docker-build
make docker-run
```

---

## TESTING LOCAL

### 1. Levantar Infraestructura
```bash
cd FarmaNexo/Helpers
./start-local.sh --full
./init-localstack-resources.sh
```

### 2. Verificar Servicios
```bash
# PostgreSQL
docker exec -it farmanexo-postgres psql -U admin -d auth_db

# Redis
docker exec -it farmanexo-redis redis-cli -a farmanexo2026

# LocalStack SQS
aws --endpoint-url=http://localhost:4566 sqs list-queues
```

### 3. Crear Base de Datos
```bash
docker exec -it farmanexo-postgres psql -U admin -c "CREATE DATABASE auth_db;"
```

### 4. Ejecutar Migraciones
```bash
cd services/auth-service
make migrate-up
```

### 5. Ejecutar Servicio
```bash
make dev
```

---

## MONITOREO

### Metricas Importantes
- Tasa de registro de usuarios
- Tasa de login exitoso/fallido
- Tasa de rate limit alcanzado
- Latencia de endpoints
- Tokens en blacklist (Redis keys count)

### Logs
- **Formato:** Console (local/dev), JSON (produccion)
- **Logger:** Zap structured logging
- **Campos contextuales:** user_id, email, correlation_id, request_id, jti

### Alertas Recomendadas
- Rate de errores 5xx > 1%
- Latencia p99 > 2s
- Redis no disponible
- SQS no disponible
- Tasa de login fallido > 50%

---

## TROUBLESHOOTING

### Problema: "Redis connection refused"
**Sintoma:** Warnings en logs sobre Redis no disponible
**Causa:** Redis no esta corriendo o password incorrecto
**Solucion:**
```bash
docker exec -it farmanexo-redis redis-cli -a farmanexo2026 ping
# Debe responder PONG
```

### Problema: "Migration failed"
**Sintoma:** Error al ejecutar `make migrate-up`
**Causa:** Schema `auth` no existe o database `auth_db` no creada
**Solucion:**
```bash
docker exec -it farmanexo-postgres psql -U admin -c "CREATE DATABASE auth_db;"
make migrate-up
```

### Problema: "SQS events not publishing"
**Sintoma:** Eventos no llegan a la cola
**Causa:** LocalStack no esta corriendo o la cola no fue creada
**Solucion:**
```bash
aws --endpoint-url=http://localhost:4566 sqs list-queues
# Si la cola no existe:
cd FarmaNexo/Helpers && ./init-localstack-resources.sh
```

### Problema: "JWT secret too short"
**Sintoma:** Panic al iniciar el servicio
**Causa:** JWT secret en config tiene menos de 32 caracteres
**Solucion:** Actualizar `jwt.secret` en config con un string de al menos 32 caracteres

---

## Referencias

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/documentation)
- [AWS SQS Documentation](https://docs.aws.amazon.com/sqs/)
- [LocalStack Documentation](https://docs.localstack.cloud/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [GORM Documentation](https://gorm.io/docs/)

---

Ultima actualizacion: 2026-02-22
