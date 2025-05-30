package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/spf13/viper"
)

const (
	Scheme        = "http"
	SchemeSecure  = "https"
	GitHub        = "github"
	StackOverflow = "stackoverflow"
	Sync          = "sync"
	Async         = "async"
	Orm           = "orm"
	Sql           = "sql"
)

type (
	Serving struct {
		ScrapperHost string `yaml:"scrapperhost" env:"SCRAPPER_HOST"`
		BotHost      string `yaml:"bothost" env:"BOT_HOST"`
		ScrapperPort string `yaml:"scrapperport" env:"SCRAPPER_PORT"`
		BotPort      string `yaml:"botport" env:"BOT_PORT"`
	}

	Database struct {
		Host     string `yaml:"host" env:"POSTGRES_HOST" envDefault:"database"`
		Port     int    `yaml:"port" env:"POSTGRES_PORT" envDefault:"5432"`
		Username string `yaml:"user" env:"POSTGRES_USER" envDefault:"postgres"`
		Password string `yaml:"password" env:"POSTGRES_PASSWORD" envDefault:"postgres"`
		Name     string `yaml:"db" env:"POSTGRES_DB" envDefault:"db"`
		Access   string `yaml:"access" envDefault:"orm"`
	}

	Secrets struct {
		GitHubToken        string `env:"GITHUB_TOKEN"`
		StackOverflowToken string `env:"STACKOVERFLOW_TOKEN"`
		BotToken           string `env:"BOT_TOKEN"`
	}

	Communication struct {
		Mode string `yaml:"mode" envDefault:"sync"`
	}
)

type Config struct {
	Serving       Serving       `yaml:"serving"`
	Database      Database      `yaml:"database"`
	Communication Communication `yaml:"communication"`
	Secrets       Secrets
}

func New(name string) (*Config, error) {
	cfg, err := NewConfigFromFile(name)
	if err != nil {
		return nil, err
	}

	if err := NewConfigFromEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func NewConfigFromFile(name string) (*Config, error) {
	cfg := &Config{}

	v := viper.New()

	v.SetConfigType("yaml")

	v.SetConfigFile(name)

	if err := v.ReadConfig(bytes.NewBuffer(configBytes)); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := v.MergeInConfig(); err != nil {
		if errors.Is(err, &viper.ConfigParseError{}) {
			return nil, fmt.Errorf("merge config: %w", err)
		}
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

func NewConfigFromEnv(cfg *Config) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	return nil
}

func (d *Database) ToDSN() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?target_session_attrs=read-write&sslmode=disable",
		d.Username,
		d.Password,
		d.Host,
		d.Port,
		d.Name,
	)
}

var (
	//go:embed default-config.yaml
	configBytes []byte
)

var Descriptions = []struct {
	Name        string
	Description string
}{
	{
		Name:        "/start",
		Description: "registers the user",
	},
	{
		Name:        "/help",
		Description: "prints the list of available commands with description",
	},
	{
		Name:        "/track",
		Description: "starts the process of adding the link",
	},
	{
		Name:        "/untrack",
		Description: "untracks the link",
	},
	{
		Name:        "/list",
		Description: "prints the list of a tracking links",
	},
	{
		Name:        "/cancel",
		Description: "return the user to the menu",
	},
}
