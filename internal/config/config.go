package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the master configuration schema for EdgeCore API Gateway
type Config struct {
	Env          string             `mapstructure:"env"`
	Gateway      GatewayConfig      `mapstructure:"gateway"`
	ControlPlane ControlPlaneConfig `mapstructure:"control_plane"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Logger       LoggerConfig       `mapstructure:"logger"`
}

// GatewayConfig contains Data Plane specific settings
type GatewayConfig struct {
	HTTPPort          int           `mapstructure:"http_port"`
	HTTPSPort         int           `mapstructure:"https_port"`
	MetricsPort       int           `mapstructure:"metrics_port"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes    int           `mapstructure:"max_header_bytes"`
	NodeID            string        `mapstructure:"node_id"`
	EnableWebSockets  bool          `mapstructure:"enable_websockets"`
	EnableCompression bool          `mapstructure:"enable_compression"`
}

// ControlPlaneConfig contains Control Plane API settings
type ControlPlaneConfig struct {
	Port         int           `mapstructure:"port"`
	JWTSecret    string        `mapstructure:"jwt_secret"`
	JWTExpiration time.Duration `mapstructure:"jwt_expiration"`
	AdminUsername string        `mapstructure:"admin_username"`
	AdminPassword string        `mapstructure:"admin_password"`
}

// DatabaseConfig contains PostgreSQL connection settings
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

// RedisConfig contains Redis cluster/standalone connection settings
type RedisConfig struct {
	Addr         string        `mapstructure:"addr"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PubSubChannel string       `mapstructure:"pubsub_channel"`
}

// LoggerConfig contains Zap logging preferences
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // "json" or "console"
	Development bool   `mapstructure:"development"`
}

// LoadConfig loads configuration from config files and environment variables
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if configPath != "" {
		v.AddConfigPath(configPath)
	}
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")

	v.SetEnvPrefix("EDGECORE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set sensible defaults
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; using defaults & env variables
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("env", "development")

	// Gateway Defaults
	v.SetDefault("gateway.http_port", 8080)
	v.SetDefault("gateway.https_port", 8443)
	v.SetDefault("gateway.metrics_port", 9090)
	v.SetDefault("gateway.read_timeout", "15s")
	v.SetDefault("gateway.write_timeout", "15s")
	v.SetDefault("gateway.idle_timeout", "60s")
	v.SetDefault("gateway.max_header_bytes", 1048576) // 1MB
	v.SetDefault("gateway.node_id", "gateway-node-1")
	v.SetDefault("gateway.enable_websockets", true)
	v.SetDefault("gateway.enable_compression", true)

	// Control Plane Defaults
	v.SetDefault("control_plane.port", 8081)
	v.SetDefault("control_plane.jwt_secret", "super-secret-jwt-key-change-in-production")
	v.SetDefault("control_plane.jwt_expiration", "24h")
	v.SetDefault("control_plane.admin_username", "admin")
	v.SetDefault("control_plane.admin_password", "admin123")

	// Database Defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.dbname", "edgecore")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", "15m")

	// Redis Defaults
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 50)
	v.SetDefault("redis.min_idle_conns", 10)
	v.SetDefault("redis.dial_timeout", "5s")
	v.SetDefault("redis.read_timeout", "3s")
	v.SetDefault("redis.write_timeout", "3s")
	v.SetDefault("redis.pubsub_channel", "edgecore:config:events")

	// Logger Defaults
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.development", true)
}
