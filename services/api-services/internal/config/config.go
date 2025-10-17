package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

var (
	cfg  *Config
	once sync.Once
)

// App modes
const (
	AppModeDevelopment = "development"
	AppModeProduction  = "production"
)

// Config is the top-level struct that holds all configuration for the application.
type Config struct {
	App          AppConfig          `envconfig:"APP"`
	Postgres     PostgresConfig     `envconfig:"POSTGRES"`
	Redis        RedisConfig        `envconfig:"REDIS"`
	ClickHouse   ClickHouseConfig   `envconfig:"CLICKHOUSE"`
	Email        EmailConfig        `envconfig:"EMAIL"`
	LocalStorage LocalStorageConfig `envconfig:"LOCAL_STORAGE"`
	Logging      LoggingConfig      `envconfig:"LOG"`
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name          string        `envconfig:"NAME" default:"UptimeApplication"`
	Key           string        `envconfig:"KEY" required:"true"`
	Port          string        `envconfig:"PORT" required:"true" default:"5005"`
	Mode          string        `envconfig:"ENV" default:"development"`
	FrontendURL   string        `envconfig:"FRONTEND_URL"`
	JWTExpiration time.Duration `envconfig:"JWT_EXPIRATION" default:"1h"`
	Version       string        `envconfig:"VERSION" default:"1.0.0"`
}

// PostgresConfig holds the configuration for the PostgreSQL database connection.
type PostgresConfig struct {
	Enable   bool   `envconfig:"ENABLE" default:"true"`
	Host     string `envconfig:"HOST" required:"true"`
	User     string `envconfig:"USERNAME" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
	Name     string `envconfig:"DATABASE" required:"true"`
	Port     int    `envconfig:"PORT" default:"5432"`
	SSLMode  string `envconfig:"SSL_MODE" default:"disable"`
}

// RedisConfig holds the configuration for the Redis connection.
type RedisConfig struct {
	Enable       bool          `envconfig:"ENABLE" default:"true"`
	Host         string        `envconfig:"HOST" default:"127.0.0.1"`
	Port         int           `envconfig:"PORT" default:"6379"`
	Password     string        `envconfig:"PASSWORD" default:""`
	DB           int           `envconfig:"DB" default:"0"`
	PoolSize     int           `envconfig:"POOL_SIZE" default:"100"`
	MinIdleConns int           `envconfig:"MIN_IDLE_CONNS" default:"10"`
	MaxConnAge   time.Duration `envconfig:"MAX_CONN_AGE" default:"1h"`
}

// ClickHouseConfig holds the configuration for the ClickHouse database connection.
type ClickHouseConfig struct {
	Enable   bool   `envconfig:"ENABLE" default:"false"`
	Host     string `envconfig:"HOST" default:"127.0.0.1"`
	Port     int    `envconfig:"PORT" default:"9000"`
	Username string `envconfig:"USERNAME" default:""`
	Password string `envconfig:"PASSWORD" default:""`
	Database string `envconfig:"DATABASE" default:"default"`
	Secure   bool   `envconfig:"SECURE" default:"false"`

	MaxIdleConns       int           `envconfig:"MAX_IDLE_CONNS" default:"10"`
	MaxOpenConns       int           `envconfig:"MAX_OPEN_CONNS" default:"50"`
	ConnMaxLifetime    time.Duration `envconfig:"CONN_MAX_LIFETIME" default:"30m"`
	ConnMaxIdleTime    time.Duration `envconfig:"CONN_MAX_IDLE_TIME" default:"5m"`
	ConnectTimeout     time.Duration `envconfig:"CONNECT_TIMEOUT" default:"5s"`
	HealthCheckTimeout time.Duration `envconfig:"HEALTH_CHECK_TIMEOUT" default:"3s"`
	MaxRetries         int           `envconfig:"MAX_RETRIES" default:"5"`
	RetryInterval      time.Duration `envconfig:"RETRY_INTERVAL" default:"2s"`
}

// EmailConfig holds the configuration for email services.
type EmailConfig struct {
	Enable             bool       `envconfig:"ENABLE" default:"false"`
	DefaultFromAddress string     `envconfig:"DEFAULT_FROM_ADDRESS" default:"no-reply@example.com"`
	DefaultProvider    string     `envconfig:"DEFAULT_PROVIDER" default:""`
	ProviderOrder      string     `envconfig:"PROVIDER_ORDER" default:""`
	SMTP               SMTPConfig `envconfig:"SMTP"`
}

// SMTPConfig holds SMTP-specific configuration.
type SMTPConfig struct {
	Enable      bool   `envconfig:"ENABLE" default:"false"`
	Host        string `envconfig:"HOST"`
	Port        int    `envconfig:"PORT"`
	Username    string `envconfig:"USERNAME"`
	Password    string `envconfig:"PASSWORD"`
	FromAddress string `envconfig:"FROM_ADDRESS"`
}

// LocalStorageConfig holds configuration for local file storage.
type LocalStorageConfig struct {
	Enable  bool   `envconfig:"ENABLE" default:"true"`
	Path    string `envconfig:"PATH" default:"./local_storage"`
	BaseURL string `envconfig:"BASE_URL" default:"/local-assets"`
}

// LoggingConfig holds logger-specific configuration
type LoggingConfig struct {
	Level       string   `envconfig:"LEVEL" default:"info"`
	Encoding    string   `envconfig:"ENCODING" default:"json"`
	Development bool     `envconfig:"DEVELOPMENT" default:"false"`
	StackTrace  bool     `envconfig:"STACKTRACE" default:"true"`
	Caller      bool     `envconfig:"CALLER" default:"true"`
	OutputPaths []string `envconfig:"OUTPUT_PATHS" default:"stdout"`

	MaxSize    int  `envconfig:"MAX_SIZE" default:"100"`
	MaxBackups int  `envconfig:"MAX_BACKUPS" default:"3"`
	MaxAge     int  `envconfig:"MAX_AGE" default:"30"`
	Compress   bool `envconfig:"COMPRESS" default:"true"`
}

// DSN generates the Data Source Name for a PostgreSQL connection.
func (p *PostgresConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.Name, p.SSLMode)
}

// GetConfig initializes and returns the application configuration.
func GetConfig() (*Config, error) {
	var err error
	once.Do(func() {
		loadedConfig, loadErr := loadConfig()
		if loadErr != nil {
			err = fmt.Errorf("failed to load configuration: %w", loadErr)
			return
		}
		cfg = loadedConfig
	})
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("configuration not initialized after sync.Once.Do, this should not happen")
	}
	return cfg, nil
}

func loadConfig() (*Config, error) {
	userConfig, err := godotenv.Read(".env")
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("warning: could not read .env: %v", err)
	}

	env := ""
	if userConfig != nil {
		env = userConfig["APP_ENV"]
	}
	if env == "" {
		env = os.Getenv("APP_ENV")
	}
	if env == "" {
		env = AppModeDevelopment
	}

	if err := godotenv.Load(fmt.Sprintf(".env.%s", env)); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env.%s: %w", env, err)
	}

	if userConfig != nil {
		for key, value := range userConfig {
			_ = os.Setenv(key, value)
		}
	}

	var c Config
	if err := envconfig.Process("", &c); err != nil {
		return nil, fmt.Errorf("failed to process configuration from environment: %w", err)
	}

	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &c, nil
}

// Validate checks for complex configuration rules.
func (c *Config) Validate() error {
	switch c.App.Mode {
	case AppModeDevelopment, AppModeProduction:
	default:
		return fmt.Errorf("invalid APP_ENV: %q, must be one of '%s', or '%s'", c.App.Mode, AppModeDevelopment, AppModeProduction)
	}

	if c.Postgres.Enable {
		if err := c.Postgres.Validate(); err != nil {
			return fmt.Errorf("postgres config invalid: %w", err)
		}
	}

	if c.Redis.Enable {
		if err := c.Redis.Validate(); err != nil {
			return fmt.Errorf("redis config invalid: %w", err)
		}
	}

	if c.ClickHouse.Enable {
		if err := c.ClickHouse.Validate(); err != nil {
			return fmt.Errorf("clickhouse config invalid: %w", err)
		}
	}

	return nil
}

// Validate methods to other config structs as needed
func (p *PostgresConfig) Validate() error {
	return nil
}

// Validate RedisConfig checks if Redis configuration is valid when enabled.
func (r *RedisConfig) Validate() error {
	if r.Host == "" {
		return fmt.Errorf("redis host is required when enabled")
	}
	if r.Port <= 0 {
		return fmt.Errorf("redis port must be a positive integer when enabled")
	}
	if r.DB < 0 {
		return fmt.Errorf("redis DB index cannot be negative")
	}
	if r.PoolSize <= 0 {
		return fmt.Errorf("redis pool size must be a positive integer")
	}
	if r.MinIdleConns < 0 {
		return fmt.Errorf("redis min idle connections cannot be negative")
	}
	if r.MinIdleConns > r.PoolSize {
		return fmt.Errorf("redis min idle connections cannot be greater than pool size")
	}
	if r.MaxConnAge < 0 {
		return fmt.Errorf("redis max connection age cannot be negative")
	}
	return nil
}

// Validate ClickHouseConfig checks if ClickHouse configuration is valid when enabled.
func (ch *ClickHouseConfig) Validate() error {
	if ch.Host == "" {
		return fmt.Errorf("clickhouse host is required when enabled")
	}
	if ch.Port <= 0 {
		return fmt.Errorf("clickhouse port must be a positive integer when enabled")
	}
	if ch.Database == "" {
		return fmt.Errorf("clickhouse database is required when enabled")
	}
	return nil
}

// String implements the fmt.Stringer interface to provide a redacted version of PostgresConfig.
func (p *PostgresConfig) String() string {
	redacted := *p
	redacted.Password = "[REDACTED]"
	return fmt.Sprintf("{Enable:%t Host:%s User:%s Name:%s Port:%d SSLMode:%s}",
		redacted.Enable, redacted.Host, redacted.User, redacted.Name, redacted.Port, redacted.SSLMode)
}

// String implements the fmt.Stringer interface to provide a redacted version of RedisConfig.
func (r *RedisConfig) String() string {
	redacted := *r
	redacted.Password = "[REDACTED]"
	return fmt.Sprintf("{Enable:%t Host:%s Port:%d DB:%d PoolSize:%d MinIdleConns:%d MaxConnAge:%v}",
		redacted.Enable, redacted.Host, redacted.Port, redacted.DB, redacted.PoolSize, redacted.MinIdleConns, redacted.MaxConnAge)
}

// DSN generates a connection string for ClickHouse.
// Format: clickhouse://username:password@host:port/database?secure=true|false
func (ch *ClickHouseConfig) DSN() string {
	secure := "false"
	if ch.Secure {
		secure = "true"
	}

	auth := ""
	if ch.Username != "" {
		auth = ch.Username
		if ch.Password != "" {
			auth += ":" + ch.Password
		}
		auth += "@"
	}

	return fmt.Sprintf("clickhouse://%s%s:%d/%s?secure=%s", auth, ch.Host, ch.Port, ch.Database, secure)
}

// String implements fmt.Stringer for ClickHouseConfig with redacted secrets.
func (ch *ClickHouseConfig) String() string {
	red := *ch
	red.Password = "[REDACTED]"
	return fmt.Sprintf("{Enable:%t Host:%s Port:%d Database:%s Username:%s Secure:%t MaxIdleConns:%d MaxOpenConns:%d}",
		red.Enable, red.Host, red.Port, red.Database, red.Username, red.Secure, red.MaxIdleConns, red.MaxOpenConns,
	)
}
