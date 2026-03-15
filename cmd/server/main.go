// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/application/handlers"
	"github.com/farmanexo/auth-service/internal/application/postprocessors"
	"github.com/farmanexo/auth-service/internal/application/preprocessors"
	"github.com/farmanexo/auth-service/internal/application/validators"
	"github.com/farmanexo/auth-service/internal/infrastructure/cache"
	"github.com/farmanexo/auth-service/internal/infrastructure/messaging"
	"github.com/farmanexo/auth-service/internal/infrastructure/persistence/postgres"
	"github.com/farmanexo/auth-service/internal/infrastructure/security"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/presentation/http/controllers"
	"github.com/farmanexo/auth-service/internal/presentation/http/middlewares"
	"github.com/farmanexo/auth-service/internal/presentation/http/routes"
	"github.com/farmanexo/auth-service/pkg/config"
	"github.com/farmanexo/auth-service/pkg/mediator"

	// Swagger docs
	_ "github.com/farmanexo/auth-service/docs"

	"go.uber.org/zap"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// @title           FarmaNexo Auth Service API
// @version         1.0
// @description     Servicio de autenticación para FarmaNexo - Microservicio con CQRS y Clean Architecture
// @termsOfService  https://farmanexo.pe/terms

// @contact.name    FarmaNexo API Support
// @contact.url     https://farmanexo.pe/support
// @contact.email   support@farmanexo.pe

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:4001
// @BasePath        /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

// @tag.name         Authentication
// @tag.description  Endpoints de autenticación

// @tag.name         Health
// @tag.description  Endpoints de salud del servicio

func main() {
	env := getEnvironment()
	cfg, err := config.LoadConfig(env)
	if err != nil {
		panic(fmt.Sprintf("Error cargando configuración: %v", err))
	}

	logger := initLogger(cfg)
	defer logger.Sync()

	logger.Info("Iniciando Auth Service",
		zap.String("environment", cfg.Environment),
		zap.Int("port", cfg.Server.Port),
	)

	db := initDatabase(cfg, logger)

	// ========================================
	// AUTO-MIGRATION DESHABILITADO
	// Usar migraciones manuales: migrate -path migrations -database "..." up
	// ========================================
	logger.Info("Auto-migration deshabilitado - Usar migraciones manuales")

	// ========================================
	// REPOSITORIOS
	// ========================================
	userRepo := postgres.NewUserRepository(db, logger)
	tokenRepo := postgres.NewTokenRepository(db, logger)

	// ========================================
	// SERVICIOS
	// ========================================
	jwtService := security.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
		cfg.JWT.Issuer,
		logger,
	)

	// ========================================
	// REDIS
	// ========================================
	redisClient, err := cache.NewRedisClient(cfg.Redis, cfg.Environment, logger)
	if err != nil {
		logger.Fatal("Error conectando a Redis", zap.Error(err))
	}

	tokenBlacklist := cache.NewRedisTokenBlacklist(redisClient, logger)
	rateLimiter := cache.NewRedisRateLimiter(redisClient, logger)

	// ========================================
	// SQS EVENT PUBLISHER
	// ========================================
	eventPublisher, err := messaging.NewSQSEventPublisher(cfg.AWS, cfg.SQS, logger)
	if err != nil {
		logger.Fatal("Error inicializando SQS EventPublisher", zap.Error(err))
	}

	// ========================================
	// MEDIATOR
	// ========================================
	med := mediator.NewMediator()

	// ========================================
	// HANDLERS
	// ========================================

	// Register Handler
	registerUserHandler := handlers.NewRegisterUserHandler(
		userRepo,
		eventPublisher,
		logger,
	)
	mediator.RegisterHandler(med, registerUserHandler)

	// Login Handler
	loginHandler := handlers.NewLoginHandler(
		userRepo,
		tokenRepo,
		jwtService,
		eventPublisher,
		rateLimiter,
		logger,
	)
	mediator.RegisterHandler(med, loginHandler)

	// Refresh Token Handler
	refreshTokenHandler := handlers.NewRefreshTokenHandler(
		userRepo,
		tokenRepo,
		jwtService,
		logger,
	)
	mediator.RegisterHandler(med, refreshTokenHandler)

	// Logout Handler
	logoutHandler := handlers.NewLogoutHandler(
		tokenRepo,
		tokenBlacklist,
		jwtService,
		eventPublisher,
		logger,
	)
	mediator.RegisterHandler(med, logoutHandler)

	// ========================================
	// VALIDATORS
	// ========================================
	registerUserValidator := validators.NewRegisterUserValidator()
	mediator.RegisterValidator[commands.RegisterUserCommand, responses.RegisterResponse](med, registerUserValidator)

	// Login Validator
	loginValidator := validators.NewLoginValidator()
	mediator.RegisterValidator[commands.LoginCommand, responses.LoginResponse](med, loginValidator)

	// ========================================
	// PREPROCESSORS Y POSTPROCESSORS
	// ========================================
	sanitizePreProcessor := preprocessors.NewSanitizeInputPreProcessor(logger)
	med.RegisterPreProcessor(sanitizePreProcessor)

	auditPostProcessor := postprocessors.NewLogAuditPostProcessor(logger)
	med.RegisterPostProcessor(auditPostProcessor)

	logger.Info("Mediator configurado",
		zap.Int("handlers", 4),
		zap.Int("validators", 2),
		zap.Int("preprocessors", 1),
		zap.Int("postprocessors", 1),
	)

	// ========================================
	// MIDDLEWARES
	// ========================================
	authMiddleware := middlewares.NewAuthMiddleware(jwtService, tokenBlacklist, logger)
	rateLimitMiddleware := middlewares.NewRateLimitMiddleware(rateLimiter, jwtService, logger)

	// ========================================
	// CONTROLADORES Y RUTAS
	// ========================================
	authController := controllers.NewAuthController(med, logger)
	router := routes.SetupRoutes(authController, authMiddleware, rateLimitMiddleware)

	// ========================================
	// SERVIDOR HTTP
	// ========================================
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("Servidor HTTP iniciado",
			zap.String("address", server.Addr),
			zap.String("swagger_url", fmt.Sprintf("http://localhost:%d/swagger/index.html", cfg.Server.Port)),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error iniciando servidor", zap.Error(err))
		}
	}()

	// ========================================
	// GRACEFUL SHUTDOWN
	// ========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Iniciando graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Error en shutdown", zap.Error(err))
	}

	// Cerrar conexión Redis
	if err := redisClient.Close(); err != nil {
		logger.Error("Error cerrando conexión Redis", zap.Error(err))
	}

	logger.Info("Servidor detenido exitosamente")
}

func getEnvironment() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}
	return env
}

func initLogger(cfg *config.Config) *zap.Logger {
	var logger *zap.Logger
	var err error

	if cfg.IsProduction() || cfg.IsUAT() {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(fmt.Sprintf("Error inicializando logger: %v", err))
	}

	return logger
}

func initDatabase(cfg *config.Config, logger *zap.Logger) *gorm.DB {
	gormLogLevel := gormlogger.Silent
	if cfg.IsDevelopment() {
		gormLogLevel = gormlogger.Info
	}

	gormLogger := gormlogger.Default.LogMode(gormLogLevel)

	db, err := gorm.Open(pgdriver.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		logger.Fatal("Error conectando a PostgreSQL",
			zap.Error(err),
			zap.String("host", cfg.Database.Host),
			zap.Int("port", cfg.Database.Port),
		)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Error obteniendo SQL DB", zap.Error(err))
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	logger.Info("Conexión a PostgreSQL establecida",
		zap.String("host", cfg.Database.Host),
		zap.String("database", cfg.Database.DBName),
	)

	return db
}
