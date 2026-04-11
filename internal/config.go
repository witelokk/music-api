package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HttpServer struct {
		Host     string `yaml:"host" env:"HOST" env-default:"0.0.0.0"`
		Port     string `yaml:"port" env:"PORT" env-default:"8080"`
		Timeouts struct {
			Read  time.Duration `yaml:"read" env:"TIMEOUT_READ" env-default:"5s"`
			Write time.Duration `yaml:"write" env:"TIMEOUT_WRITE" env-default:"5s"`
			Idle  time.Duration `yaml:"idle" env:"TIMEOUT_IDLE" env-default:"5s"`
		}
	} `yaml:"http_server"`

	Auth struct {
		JWTSecret                   string        `yaml:"jwt_secret" env:"JWT_SECRET" env-required:"true"`
		AccessTokenTTL              time.Duration `yaml:"access_token_ttl" env:"ACCESS_TOKEN_TTL" env-default:"15m"`
		RefreshTokenTTL             time.Duration `yaml:"refresh_token_ttl" env:"REFRESH_TOKEN_TTL" env-default:"168h"` // 7 days
		VerificationCodeTTL         time.Duration `yaml:"verification_code_ttl" env:"VERIFICATION_CODE_TTL" env-default:"15m"`
		NewVerificationCodeInterval time.Duration `yaml:"new_verification_code_interval" env:"NEW_VERIFICATION_CODE_INTERVAL" env-default:"2m"`
		GoogleIdTokenAudiences      []string      `yaml:"google_id_token_audiences" env-delim:"," env:"GOOGLE_ID_TOKEN_AUDIENCES" env-required:"true"`
	} `yaml:"auth"`

	DatabaseURL string `env:"DATABASE_URL" yaml:"database.url" env-required:"true"`
	RedisURL    string `env:"REDIS_URL" yaml:"redis.url" env-required:"true"`

	Mailgun struct {
		APIKey string `env:"MAILGUN_API_KEY" yaml:"mailgun.api_key" env-required:"true"`
		Domain string `env:"MAILGUN_DOMAIN" yaml:"mailgun.domain" env-required:"true"`
		From   string `env:"MAILGUN_FROM" yaml:"mailgun.from" env-required:"true"`
		Region string `env:"MAILGUN_REGION" yaml:"mailgun.region" env-required:"true"`
	} `yaml:"mailgun"`

	Minio struct {
		Endpoint  string `env:"MINIO_ENDPOINT" yaml:"minio.endpoint" env-required:"true"`
		AccessKey string `env:"MINIO_ACCESS_KEY" yaml:"minio.access_key" env-required:"true"`
		SecretKey string `env:"MINIO_SECRET_KEY" yaml:"minio.secret_key" env-required:"true"`
		Bucket    string `env:"MINIO_BUCKET" yaml:"minio.bucket" env-required:"true"`
		UseSSL    bool   `env:"MINIO_USE_SSL" yaml:"minio.use_ssl" env-default:"false"`
	} `yaml:"minio"`

	Logger struct {
		Type  string `yaml:"type" env:"LOGGER_TYPE" env-default:"text"`
		Level string `yaml:"level" env:"LOGGER_LEVEL" env-default:"info"`
	}
}

func LoadConfig() (*Config, error) {
	cfg := new(Config)

	if configPath, ok := os.LookupEnv("CONFIG_PATH"); ok {
		if _, err := os.Stat(configPath); err != nil {
			return nil, err
		}
		if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}

	switch cfg.Mailgun.Region {
	case "EU", "US":
	default:
		return nil, fmt.Errorf("invalid Mailgun region %q, must be \"EU\" or \"US\"", cfg.Mailgun.Region)
	}

	return cfg, nil
}

func MustLoadConfig() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}
