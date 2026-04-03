package internal

import (
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
	}

	Database struct {
		URL      string `env:"DATABASE_URL" yaml:"database.url"`
		User     string `env:"DATABASE_USER" yaml:"database.user"`
		Password string `env:"DATABASE_PASSWORD" yaml:"database.password"`
	} `yaml:"database"`

	RedisURL  string `env:"REDIS_URL" yaml:"redis.url"`
	JWTSecret string `env:"JWT_SECRET" yaml:"jwt.secret"`

	Mailgun struct {
		APIKey string `env:"MAILGUN_API_KEY" yaml:"mailgun.api_key"`
		Domain string `env:"MAILGUN_DOMAIN" yaml:"mailgun.domain"`
		From   string `env:"MAILGUN_FROM" yaml:"mailgun.from"`
		Region string `env:"MAILGUN_REGION" yaml:"mailgun.region"`
	} `yaml:"mailgun"`

	GoogleAuth struct {
		Audience []string `env:"GOOGLE_AUTH_AUDIENCE" env-delim:"," yaml:"google-auth.audience"`
	} `yaml:"google-auth"`

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

	return cfg, nil
}

func MustLoadConfig() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}
