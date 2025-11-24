package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	TOTP         TOTPConfig         `mapstructure:"totp"`
	Storage      StorageConfig      `mapstructure:"storage"`
	SystemPolicy SystemPolicyConfig `mapstructure:"system_policy"`
	CORS         CORSConfig         `mapstructure:"cors"`
	RateLimit    RateLimitConfig    `mapstructure:"rate_limit"`
	Logging      LoggingConfig      `mapstructure:"logging"`
	Cleanup      CleanupConfig      `mapstructure:"cleanup"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Email        EmailConfig        `mapstructure:"email"`
	CloudStorage CloudStorageConfig `mapstructure:"cloud_storage"`
	Metrics      MetricsConfig      `mapstructure:"metrics"`
	Swagger      SwaggerConfig      `mapstructure:"swagger"`
}

type ServerConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Mode            string `mapstructure:"mode"`
	ReadTimeout     string `mapstructure:"read_timeout"`
	WriteTimeout    string `mapstructure:"write_timeout"`
	ShutdownTimeout string `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"sslmode"`
	Timezone        string `mapstructure:"timezone"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
}

type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	AccessTokenExpiry  string `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry string `mapstructure:"refresh_token_expiry"`
}

type TOTPConfig struct {
	Issuer string `mapstructure:"issuer"`
	Period uint   `mapstructure:"period"`
	Digits uint   `mapstructure:"digits"`
}

type StorageConfig struct {
	Type             string   `mapstructure:"type"`
	Path             string   `mapstructure:"path"`
	MaxFileSizeMB    int      `mapstructure:"max_file_size_mb"`
	AllowedMimeTypes []string `mapstructure:"allowed_mime_types"`
}

type CloudStorageConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	Provider         string `mapstructure:"provider"` // e.g. "azure"
	Endpoint         string `mapstructure:"endpoint"`
	AccessKey        string `mapstructure:"access_key"`          // Azure: Storage Account Name
	SecretKey        string `mapstructure:"secret_key"`          // Azure: Storage Account Key / SAS
	PublicContainer  string `mapstructure:"public_container"`    // Azure: public container name
	PrivateContainer string `mapstructure:"private_container"`   // Azure: private container name
	Region           string `mapstructure:"region"`              // Azure: optional, keep for consistency
}

type SystemPolicyConfig struct {
	DefaultValidityDays int `mapstructure:"default_validity_days"`
	MinValidityHours    int `mapstructure:"min_validity_hours"`
	MaxValidityDays     int `mapstructure:"max_validity_days"`
	MinPasswordLength   int `mapstructure:"min_password_length"`
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           string   `mapstructure:"max_age"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	UploadPerHour     int `mapstructure:"upload_per_hour"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type CleanupConfig struct {
	Cron   string `mapstructure:"cron"`
	Secret string `mapstructure:"secret"`
}

type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type EmailConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	SMTPHost string `mapstructure:"smtp_host"`
	SMTPPort int    `mapstructure:"smtp_port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type MetricsConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
}

type SwaggerConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	BasePath string `mapstructure:"base_path"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	// Read environment variables
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with environment variables
	if port := os.Getenv("PORT"); port != "" {
		v.Set("server.port", port)
	}
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cfg.Database.Host = dbHost
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		cfg.Database.Password = dbPass
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.JWT.Secret = jwtSecret
	}
	
	// Override cloud storage settings from environment
	if enabled := os.Getenv("CLOUD_STORAGE_ENABLED"); enabled != "" {
		cfg.CloudStorage.Enabled = enabled == "true"
	}
	if provider := os.Getenv("CLOUD_STORAGE_PROVIDER"); provider != "" {
		cfg.CloudStorage.Provider = provider
	}
	if endpoint := os.Getenv("CLOUD_STORAGE_ENDPOINT"); endpoint != "" {
		cfg.CloudStorage.Endpoint = endpoint
	}
	if accessKey := os.Getenv("CLOUD_STORAGE_ACCESS_KEY"); accessKey != "" {
		cfg.CloudStorage.AccessKey = accessKey
	}
	if secretKey := os.Getenv("CLOUD_STORAGE_SECRET_KEY"); secretKey != "" {
		cfg.CloudStorage.SecretKey = secretKey
	}
	if publicContainer := os.Getenv("CLOUD_STORAGE_PUBLIC_CONTAINER"); publicContainer != "" {
		cfg.CloudStorage.PublicContainer = publicContainer
	}
	if privateContainer := os.Getenv("CLOUD_STORAGE_PRIVATE_CONTAINER"); privateContainer != "" {
		cfg.CloudStorage.PrivateContainer = privateContainer
	}

	return &cfg, nil
}

// GetDSN returns PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode, c.Timezone,
	)
}

// Helper methods to parse duration strings
func (c *JWTConfig) GetAccessTokenExpiry() (time.Duration, error) {
	return parseDuration(c.AccessTokenExpiry)
}

func (c *JWTConfig) GetRefreshTokenExpiry() (time.Duration, error) {
	return parseDuration(c.RefreshTokenExpiry)
}

func (c *ServerConfig) GetReadTimeout() (time.Duration, error) {
	return parseDuration(c.ReadTimeout)
}

func (c *ServerConfig) GetWriteTimeout() (time.Duration, error) {
	return parseDuration(c.WriteTimeout)
}

func (c *ServerConfig) GetShutdownTimeout() (time.Duration, error) {
	return parseDuration(c.ShutdownTimeout)
}

func (c *DatabaseConfig) GetConnMaxLifetime() (time.Duration, error) {
	return parseDuration(c.ConnMaxLifetime)
}

func (c *CORSConfig) GetMaxAge() (time.Duration, error) {
	return parseDuration(c.MaxAge)
}

// parseDuration parses duration strings like "7d", "24h", "30m"
func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}

	// Handle days (e.g., "7d")
	if len(s) > 1 && s[len(s)-1] == 'd' {
		days := s[:len(s)-1]
		var d int
		_, err := fmt.Sscanf(days, "%d", &d)
		if err != nil {
			return 0, fmt.Errorf("invalid duration format: %s", s)
		}
		return time.Duration(d) * 24 * time.Hour, nil
	}

	// Use standard time.ParseDuration for other formats
	return time.ParseDuration(s)
}
