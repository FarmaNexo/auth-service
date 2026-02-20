# 🏗️ INFRAESTRUCTURA LOCAL - AUTH SERVICE

## 📊 Resumen

Este documento describe la infraestructura local disponible y cómo integrarla con el Auth Service.

---

## 🐳 SERVICIOS DISPONIBLES

### Base de Datos

- **PostgreSQL 16 (PostGIS)**: `localhost:5432`
  - User: `admin`
  - Password: `admin`
  - Database: `auth_db`
  - Schema: `auth`
  - Adminer UI: <http://localhost:8082>

### Cache

- **Redis 7**: `localhost:6379`
  - Password: `farmanexo2026`
  - Redis Commander UI: <http://localhost:8081>

### AWS LocalStack (Simulación AWS)

- **Gateway**: `http://localhost:4566`
- **Health**: `http://localhost:4566/_localstack/health`
- **Credenciales** (fake):
  - AWS_ACCESS_KEY_ID: `test`
  - AWS_SECRET_ACCESS_KEY: `test`
  - AWS_DEFAULT_REGION: `us-east-1`

### Monitoring

- **Prometheus**: <http://localhost:9090>
- **Grafana**: <http://localhost:3001> (admin/admin)

---

## ☁️ RECURSOS AWS EN LOCALSTACK

### S3 Buckets

```
farmanexo-products       # Imágenes de productos
farmanexo-pharmacies     # Logos de farmacias
farmanexo-avatars        # Avatares de usuarios
farmanexo-documents      # Documentos varios
```

### SQS Queues (Reemplazo de Kafka)

```
farmanexo-auth-events         # Eventos de autenticación
farmanexo-catalog-events      # Eventos de catálogo
farmanexo-pharmacy-events     # Eventos de farmacias
farmanexo-price-events        # Eventos de precios
farmanexo-user-events         # Eventos de usuarios
farmanexo-analytics-events    # Eventos de analytics
farmanexo-notifications       # Notificaciones
farmanexo-dlq                 # Dead Letter Queue
```

### SNS Topics (Pub/Sub)

```
farmanexo-user-registered     # Usuario registrado
farmanexo-product-created     # Producto creado
farmanexo-price-updated       # Precio actualizado
farmanexo-order-placed        # Orden colocada
```

### Secrets Manager

```
farmanexo/auth/jwt-secret           # JWT secret key
farmanexo/database/password         # Database password
```

### Parameter Store (SSM)

```
/farmanexo/auth/database-host       # localhost
/farmanexo/auth/redis-host          # localhost
/farmanexo/environment              # local
```

### DynamoDB

```
farmanexo-sessions                  # Sesiones de usuario
```

---

## 🔧 CÓMO USAR REDIS

### 1. Configuración en config.local.yaml

```yaml
redis:
  host: localhost
  port: 6379
  password: farmanexo2026
  db: 0
  max_retries: 3
  pool_size: 10
```

### 2. Cliente Redis (interno/infrastructure/cache/redis_client.go)

```go
package cache

import (
    "context"
    "time"
    "github.com/redis/go-redis/v9"
)

type RedisClient struct {
    client *redis.Client
}

func NewRedisClient(host, password string, db int) (*RedisClient, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     host + ":6379",
        Password: password,
        DB:       db,
    })
    
    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, err
    }
    
    return &RedisClient{client: client}, nil
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
    return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Delete(ctx context.Context, key string) error {
    return r.client.Del(ctx, key).Err()
}
```

### 3. Casos de Uso

- Cache de tokens JWT (blacklist para logout)
- Rate limiting por usuario/IP
- Sesiones temporales
- Cache de queries frecuentes

---

## 🚀 CÓMO USAR SQS (Reemplazo de Kafka)

### 1. Cliente SQS (internal/infrastructure/messaging/sqs_client.go)

```go
package messaging

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSClient struct {
    client *sqs.Client
}

func NewSQSClient(ctx context.Context) (*SQSClient, error) {
    cfg, err := config.LoadDefaultConfig(ctx,
        config.WithRegion("us-east-1"),
        config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
            func(service, region string, options ...interface{}) (aws.Endpoint, error) {
                // Para ambiente local, usar LocalStack
                if os.Getenv("ENV") == "local" {
                    return aws.Endpoint{
                        URL: "http://localhost:4566",
                    }, nil
                }
                return aws.Endpoint{}, &aws.EndpointNotFoundError{}
            },
        )),
    )
    if err != nil {
        return nil, err
    }
    
    return &SQSClient{
        client: sqs.NewFromConfig(cfg),
    }, nil
}

func (s *SQSClient) SendMessage(ctx context.Context, queueURL, message string) error {
    _, err := s.client.SendMessage(ctx, &sqs.SendMessageInput{
        QueueUrl:    aws.String(queueURL),
        MessageBody: aws.String(message),
    })
    return err
}
```

### 2. Publicar Eventos

```go
// Evento: Usuario registrado
event := map[string]interface{}{
    "event_type": "USER_REGISTERED",
    "user_id":    user.ID,
    "email":      user.Email,
    "timestamp":  time.Now(),
}

eventJSON, _ := json.Marshal(event)
sqsClient.SendMessage(ctx, "http://localhost:4566/000000000000/farmanexo-auth-events", string(eventJSON))
```

---

## 📦 CÓMO USAR S3

### 1. Cliente S3 (internal/infrastructure/storage/s3_client.go)

```go
package storage

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
    client *s3.Client
}

func (s *S3Client) Upload(ctx context.Context, bucket, key string, data []byte) error {
    _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
        Body:   bytes.NewReader(data),
    })
    return err
}
```

### 2. Casos de Uso

- Subir avatares de usuarios: `farmanexo-avatars/users/{user_id}.jpg`
- Documentos de verificación: `farmanexo-documents/kyc/{user_id}/...`

---

## 🔐 CÓMO USAR SECRETS MANAGER

### 1. Obtener Secrets

```go
import "github.com/aws/aws-sdk-go-v2/service/secretsmanager"

func GetJWTSecret(ctx context.Context) (string, error) {
    cfg, _ := config.LoadDefaultConfig(ctx)
    client := secretsmanager.NewFromConfig(cfg)
    
    result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
        SecretId: aws.String("farmanexo/auth/jwt-secret"),
    })
    
    return *result.SecretString, err
}
```

---

## 🎯 INTEGRACIÓN RECOMENDADA PARA AUTH SERVICE

### Fase 1: Redis (Inmediato)

1. ✅ Implementar blacklist de tokens en logout
2. ✅ Rate limiting de login attempts
3. ✅ Cache de refresh tokens activos

### Fase 2: SQS (Siguiente)

1. ✅ Publicar evento `USER_REGISTERED` después de registro exitoso
2. ✅ Publicar evento `USER_LOGIN` en cada login
3. ✅ Publicar evento `USER_LOGOUT` en logout

### Fase 3: S3 (Futuro)

1. Upload de avatares de usuario
2. Documentos de verificación KYC

### Fase 4: Secrets Manager (Opcional)

1. Mover JWT secret de config a Secrets Manager
2. Rotar secrets automáticamente

---

## 🚀 COMANDOS ÚTILES

### Iniciar infraestructura (WSL2)

```bash
cd /mnt/c/Users/Usuario/Desktop/Proyectos/FarmaNexo/Helpers
./start-local.sh --full
./init-localstack-resources.sh
```

### Verificar servicios

```bash
# PostgreSQL
docker exec -it farmanexo-postgres psql -U admin -d auth_db

# Redis
docker exec -it farmanexo-redis redis-cli -a farmanexo2026

# LocalStack S3
awslocal s3 ls

# LocalStack SQS
awslocal sqs list-queues
```

### Migraciones y servicio

```bash
cd /mnt/c/Users/Usuario/Desktop/Proyectos/FarmaNexo/services/auth-service
make migrate-up
make dev
```

---

## 📚 DEPENDENCIAS GO NECESARIAS

```bash
# Redis
go get github.com/redis/go-redis/v9

# AWS SDK
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/service/sqs
go get github.com/aws/aws-sdk-go-v2/service/sns
go get github.com/aws/aws-sdk-go-v2/service/secretsmanager
```

---

## ⚠️ IMPORTANTE

- **Ambiente LOCAL**: Usar LocalStack (endpoint: <http://localhost:4566>)
- **Ambiente CLOUD**: Usar AWS real (sin endpoint custom)
- **Variables de entorno**: Usar ENV para detectar ambiente

```

---

## 📋 PROMPT PARA CLAUDE CODE

Copia y pega esto en Claude Code:
```

# CONTEXTO: INFRAESTRUCTURA LOCAL DISPONIBLE

Tengo la siguiente infraestructura corriendo localmente:

## Servicios activos

1. PostgreSQL (localhost:5432) - admin/admin - Database: auth_db
2. Redis (localhost:6379) - Password: farmanexo2026
3. LocalStack AWS (localhost:4566) simulando:
   - S3 (4 buckets: products, pharmacies, avatars, documents)
   - SQS (8 queues para eventos, reemplazando Kafka)
   - SNS (4 topics para pub/sub)
   - Secrets Manager (JWT secret, DB password)
   - Parameter Store (config parameters)
   - DynamoDB (tabla de sesiones)

## Proyecto actual

- Auth Service en Go con Clean Architecture + CQRS + MediatR
- Ya tiene: Register, Login, Refresh, GetProfile, Logout
- PostgreSQL ya integrado y funcionando
- Migraciones con golang-migrate funcionando

## Lo que necesito

1. Integrar Redis para:
   - Blacklist de tokens en logout
   - Rate limiting de login attempts
   - Cache de refresh tokens activos

2. Integrar SQS para publicar eventos:
   - USER_REGISTERED (después de registro)
   - USER_LOGIN (en cada login)
   - USER_LOGOUT (en logout)

3. Configuración que detecte ambiente (local usa LocalStack, cloud usa AWS real)

## Archivos importantes

- configs/config.local.yaml (configuración local)
- Ver archivo INFRASTRUCTURE.md para detalles completos

¿Puedes ayudarme a implementar primero la integración con Redis?
