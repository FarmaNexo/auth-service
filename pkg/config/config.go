// pkg/config/config.go
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config representa toda la configuración del servicio
type Config struct {
	Environment string         `mapstructure:"environment"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	JWT         JWTConfig      `mapstructure:"jwt"`
	Redis       RedisConfig    `mapstructure:"redis"`
	AWS         AWSConfig      `mapstructure:"aws"`
	SQS         SQSConfig      `mapstructure:"sqs"`
	Log         LogConfig      `mapstructure:"log"`
}

// ServerConfig configuración del servidor HTTP
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig configuración de PostgreSQL
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// JWTConfig configuración de JWT
type JWTConfig struct {
	Secret               string        `mapstructure:"secret"`
	AccessTokenDuration  time.Duration `mapstructure:"access_token_duration"`
	RefreshTokenDuration time.Duration `mapstructure:"refresh_token_duration"`
	Issuer               string        `mapstructure:"issuer"`
}

// AWSConfig configuración de AWS
type AWSConfig struct {
	Region   string `mapstructure:"region"`
	Endpoint string `mapstructure:"endpoint"` // Vacío para AWS real, URL de LocalStack para local
}

// SQSConfig configuración de SQS
type SQSConfig struct {
	AuthEventsQueueURL string `mapstructure:"auth_events_queue_url"`
}

// RedisConfig configuración de Redis
type RedisConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	MaxRetries int    `mapstructure:"max_retries"`
	PoolSize   int    `mapstructure:"pool_size"`
}

// GetAddr retorna la dirección host:port de Redis
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// LogConfig configuración de logging
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"` // json, console
}

// ========================================
// LOAD CONFIG
// ========================================

// LoadConfig carga la configuración basada en el environment
func LoadConfig(environment string) (*Config, error) {
	v := viper.New()

	// Configurar Viper
	v.SetConfigName(fmt.Sprintf("config.%s", environment))
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")

	// Variables de entorno tienen prioridad
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Leer el archivo de configuración
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal a struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Establecer environment
	config.Environment = environment

	// Validar configuración
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// ========================================
// VALIDATION
// ========================================

// validateConfig valida que la configuración sea correcta
func validateConfig(config *Config) error {
	// Validar Server
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validar Database
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	// Validar JWT
	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if len(config.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}

	return nil
}

// ========================================
// HELPERS
// ========================================

// GetDatabaseDSN retorna el Data Source Name para PostgreSQL
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.DBName,
		c.SSLMode,
	)
}

// IsProduction verifica si está en producción
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsLocal verifica si está en máquina local
func (c *Config) IsLocal() bool {
	return c.Environment == "local"
}

// IsDevelopment verifica si está en desarrollo (desplegado)
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsQA verifica si está en QA
func (c *Config) IsQA() bool {
	return c.Environment == "qa"
}

// IsUAT verifica si está en UAT
func (c *Config) IsUAT() bool {
	return c.Environment == "uat"
}
