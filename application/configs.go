package application

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cerhlhgr/golang-lib/config"
	httpPkg "github.com/cerhlhgr/golang-lib/http"
)

type HTTPConfig struct {
	Server                httpPkg.Config
	EnableHealthEndpoints bool
	HealthPath            string
	ReadinessPath         string
	LivenessPath          string
	CheckTimeout          time.Duration
}

func DefaultHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Server:                httpPkg.DefaultConfig(),
		EnableHealthEndpoints: true,
		HealthPath:            "/health",
		ReadinessPath:         "/ready",
		LivenessPath:          "/live",
		CheckTimeout:          2 * time.Second,
	}
}

func HTTPConfigFromEnv() HTTPConfig {
	cfg := DefaultHTTPConfig()
	cfg.Server.Name = config.GetString("HTTP_SERVER_NAME", cfg.Server.Name)
	cfg.Server.Addr = config.GetString("HTTP_ADDR", cfg.Server.Addr)
	cfg.Server.ReadTimeout = config.GetDuration("HTTP_READ_TIMEOUT", cfg.Server.ReadTimeout)
	cfg.Server.ReadHeaderTimeout = config.GetDuration("HTTP_READ_HEADER_TIMEOUT", cfg.Server.ReadHeaderTimeout)
	cfg.Server.WriteTimeout = config.GetDuration("HTTP_WRITE_TIMEOUT", cfg.Server.WriteTimeout)
	cfg.Server.IdleTimeout = config.GetDuration("HTTP_IDLE_TIMEOUT", cfg.Server.IdleTimeout)
	cfg.Server.ShutdownTimeout = config.GetDuration("HTTP_SHUTDOWN_TIMEOUT", cfg.Server.ShutdownTimeout)
	cfg.Server.MaxHeaderBytes = config.GetInt("HTTP_MAX_HEADER_BYTES", cfg.Server.MaxHeaderBytes)
	cfg.EnableHealthEndpoints = config.GetBool("HTTP_ENABLE_HEALTH", cfg.EnableHealthEndpoints)
	cfg.HealthPath = config.GetString("HTTP_HEALTH_PATH", cfg.HealthPath)
	cfg.ReadinessPath = config.GetString("HTTP_READINESS_PATH", cfg.ReadinessPath)
	cfg.LivenessPath = config.GetString("HTTP_LIVENESS_PATH", cfg.LivenessPath)
	cfg.CheckTimeout = config.GetDuration("HTTP_HEALTHCHECK_TIMEOUT", cfg.CheckTimeout)
	return cfg
}

func (c *HTTPConfig) setDefaults() {
	def := DefaultHTTPConfig()
	c.Server.SetDefaults()
	if c.HealthPath == "" {
		c.HealthPath = def.HealthPath
	}
	if c.ReadinessPath == "" {
		c.ReadinessPath = def.ReadinessPath
	}
	if c.LivenessPath == "" {
		c.LivenessPath = def.LivenessPath
	}
	if c.CheckTimeout <= 0 {
		c.CheckTimeout = def.CheckTimeout
	}
}

func (c HTTPConfig) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return err
	}

	if strings.TrimSpace(c.HealthPath) == "" {
		return errors.New("http config: health path is required")
	}
	if strings.TrimSpace(c.ReadinessPath) == "" {
		return errors.New("http config: readiness path is required")
	}
	if strings.TrimSpace(c.LivenessPath) == "" {
		return errors.New("http config: liveness path is required")
	}

	return nil
}

type PostgresConfig struct {
	DSN             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	HealthCheck     time.Duration
	ConnectTimeout  time.Duration
	PingOnStart     bool
	PingTimeout     time.Duration
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		MaxConns:        10,
		MinConns:        0,
		MaxConnLifetime: 60 * time.Minute,
		MaxConnIdleTime: 30 * time.Minute,
		HealthCheck:     1 * time.Minute,
		ConnectTimeout:  5 * time.Second,
		PingOnStart:     true,
		PingTimeout:     2 * time.Second,
	}
}

func PostgresConfigFromEnv() PostgresConfig {
	cfg := DefaultPostgresConfig()
	cfg.DSN = config.GetString("POSTGRES_DSN", "")
	cfg.MaxConns = int32(config.GetInt("PG_MAX_CONNS", int(cfg.MaxConns)))
	cfg.MinConns = int32(config.GetInt("PG_MIN_CONNS", int(cfg.MinConns)))
	cfg.MaxConnLifetime = config.GetDuration("PG_MAX_CONN_LIFETIME", cfg.MaxConnLifetime)
	cfg.MaxConnIdleTime = config.GetDuration("PG_MAX_CONN_IDLE_TIME", cfg.MaxConnIdleTime)
	cfg.HealthCheck = config.GetDuration("PG_HEALTH_CHECK_PERIOD", cfg.HealthCheck)
	cfg.ConnectTimeout = config.GetDuration("PG_CONNECT_TIMEOUT", cfg.ConnectTimeout)
	cfg.PingOnStart = config.GetBool("PG_PING_ON_START", cfg.PingOnStart)
	cfg.PingTimeout = config.GetDuration("PG_PING_TIMEOUT", cfg.PingTimeout)
	return cfg
}

func (c *PostgresConfig) setDefaults() {
	def := DefaultPostgresConfig()
	if c.MaxConns <= 0 {
		c.MaxConns = def.MaxConns
	}
	if c.MaxConnLifetime <= 0 {
		c.MaxConnLifetime = def.MaxConnLifetime
	}
	if c.MaxConnIdleTime <= 0 {
		c.MaxConnIdleTime = def.MaxConnIdleTime
	}
	if c.HealthCheck <= 0 {
		c.HealthCheck = def.HealthCheck
	}
	if c.ConnectTimeout <= 0 {
		c.ConnectTimeout = def.ConnectTimeout
	}
	if c.PingTimeout <= 0 {
		c.PingTimeout = def.PingTimeout
	}
}

func (c PostgresConfig) Validate() error {
	if strings.TrimSpace(c.DSN) == "" {
		return errors.New("postgres config: dsn is required")
	}
	if c.MaxConns <= 0 {
		return errors.New("postgres config: max_conns must be > 0")
	}
	if c.MinConns < 0 {
		return errors.New("postgres config: min_conns must be >= 0")
	}
	if c.MinConns > c.MaxConns {
		return fmt.Errorf("postgres config: min_conns (%d) cannot be greater than max_conns (%d)", c.MinConns, c.MaxConns)
	}
	return nil
}

type RedisConfig struct {
	Addr         string
	Username     string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	PoolTimeout  time.Duration
	PingOnStart  bool
	PingTimeout  time.Duration
}

func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     10,
		MinIdleConns: 0,
		MaxRetries:   2,
		PoolTimeout:  4 * time.Second,
		PingOnStart:  true,
		PingTimeout:  2 * time.Second,
	}
}

func RedisConfigFromEnv() RedisConfig {
	cfg := DefaultRedisConfig()
	cfg.Addr = config.GetString("REDIS_ADDR", "")
	cfg.Username = config.GetString("REDIS_USERNAME", "")
	cfg.Password = config.GetString("REDIS_PASSWORD", "")
	cfg.DB = config.GetInt("REDIS_DB", 0)
	cfg.DialTimeout = config.GetDuration("REDIS_DIAL_TIMEOUT", cfg.DialTimeout)
	cfg.ReadTimeout = config.GetDuration("REDIS_READ_TIMEOUT", cfg.ReadTimeout)
	cfg.WriteTimeout = config.GetDuration("REDIS_WRITE_TIMEOUT", cfg.WriteTimeout)
	cfg.PoolSize = config.GetInt("REDIS_POOL_SIZE", cfg.PoolSize)
	cfg.MinIdleConns = config.GetInt("REDIS_MIN_IDLE_CONNS", cfg.MinIdleConns)
	cfg.MaxRetries = config.GetInt("REDIS_MAX_RETRIES", cfg.MaxRetries)
	cfg.PoolTimeout = config.GetDuration("REDIS_POOL_TIMEOUT", cfg.PoolTimeout)
	cfg.PingOnStart = config.GetBool("REDIS_PING_ON_START", cfg.PingOnStart)
	cfg.PingTimeout = config.GetDuration("REDIS_PING_TIMEOUT", cfg.PingTimeout)
	return cfg
}

func (c *RedisConfig) setDefaults() {
	def := DefaultRedisConfig()
	if c.DialTimeout <= 0 {
		c.DialTimeout = def.DialTimeout
	}
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = def.ReadTimeout
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = def.WriteTimeout
	}
	if c.PoolSize <= 0 {
		c.PoolSize = def.PoolSize
	}
	if c.PoolTimeout <= 0 {
		c.PoolTimeout = def.PoolTimeout
	}
	if c.PingTimeout <= 0 {
		c.PingTimeout = def.PingTimeout
	}
}

func (c RedisConfig) Validate() error {
	if strings.TrimSpace(c.Addr) == "" {
		return errors.New("redis config: addr is required")
	}
	if c.PoolSize <= 0 {
		return errors.New("redis config: pool_size must be > 0")
	}
	if c.MinIdleConns < 0 {
		return errors.New("redis config: min_idle_conns must be >= 0")
	}
	if c.MinIdleConns > c.PoolSize {
		return fmt.Errorf("redis config: min_idle_conns (%d) cannot be greater than pool_size (%d)", c.MinIdleConns, c.PoolSize)
	}
	return nil
}

type S3Config struct {
	Endpoint         string
	Region           string
	AccessKey        string
	SecretKey        string
	Bucket           string
	UseSSL           bool
	ForcePathStyle   bool
	CheckBucket      bool
	AutoCreateBucket bool
}

func DefaultS3Config() S3Config {
	return S3Config{
		UseSSL:         false,
		ForcePathStyle: true,
		CheckBucket:    true,
	}
}

func S3ConfigFromEnv() S3Config {
	cfg := DefaultS3Config()
	cfg.Endpoint = config.GetString("S3_ENDPOINT", "")
	cfg.AccessKey = config.GetString("S3_ACCESS_KEY", "")
	cfg.SecretKey = config.GetString("S3_SECRET_KEY", "")
	cfg.Region = config.GetString("S3_REGION", "")
	cfg.Bucket = config.GetString("S3_BUCKET", "")
	cfg.UseSSL = config.GetBool("S3_USE_SSL", cfg.UseSSL)
	cfg.ForcePathStyle = config.GetBool("S3_FORCE_PATH_STYLE", cfg.ForcePathStyle)
	cfg.CheckBucket = config.GetBool("S3_CHECK_BUCKET", cfg.CheckBucket)
	cfg.AutoCreateBucket = config.GetBool("S3_AUTO_CREATE_BUCKET", cfg.AutoCreateBucket)
	return cfg
}

func (c S3Config) Validate() error {
	if strings.TrimSpace(c.Endpoint) == "" {
		return errors.New("s3 config: endpoint is required")
	}
	if strings.TrimSpace(c.AccessKey) == "" {
		return errors.New("s3 config: access_key is required")
	}
	if strings.TrimSpace(c.SecretKey) == "" {
		return errors.New("s3 config: secret_key is required")
	}

	return nil
}

func WithHTTP(cfg HTTPConfig) Option {
	return func(a *Application) error {
		cfg.setDefaults()
		if err := cfg.Validate(); err != nil {
			return err
		}

		a.httpConfig = cfg
		return addServiceToApp(httpServer, a)
	}
}

func WithHTTPHandler(handler http.Handler) Option {
	return func(a *Application) error {
		a.HTTP = handler
		return nil
	}
}

func WithPostgres(cfg PostgresConfig) Option {
	return func(a *Application) error {
		cfg.setDefaults()
		if err := cfg.Validate(); err != nil {
			return err
		}

		a.postgresConfig = &cfg
		if ok := a.dependencies.addDependency(Postgres); !ok {
			return fmt.Errorf("%s: %w", Postgres.name, ErrDependencyAlreadyEnabled)
		}

		return nil
	}
}

func WithRedis(cfg RedisConfig) Option {
	return func(a *Application) error {
		cfg.setDefaults()
		if err := cfg.Validate(); err != nil {
			return err
		}

		a.redisConfig = &cfg
		if ok := a.dependencies.addDependency(Redis); !ok {
			return fmt.Errorf("%s: %w", Redis.name, ErrDependencyAlreadyEnabled)
		}

		return nil
	}
}

func WithS3(cfg S3Config) Option {
	return func(a *Application) error {
		if err := cfg.Validate(); err != nil {
			return err
		}

		a.s3Config = &cfg
		if ok := a.dependencies.addDependency(S3); !ok {
			return fmt.Errorf("%s: %w", S3.name, ErrDependencyAlreadyEnabled)
		}

		return nil
	}
}
