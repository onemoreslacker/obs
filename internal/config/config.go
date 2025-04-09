package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type (
	Serving struct {
		ScrapperHost string `yaml:"scrapper-host" env:"SCRAPPER_HOST" env-default:"scrapper"`
		BotHost      string `yaml:"bot-host" env:"BOT_HOST" env-default:"bot"`
		ScrapperPort string `yaml:"scrapper_port" env:"SCRAPPER_PORT" env-default:"8080"`
		BotPort      string `yaml:"bot_port" env:"BOT_PORT" env-default:"8081"`
	}

	Database struct {
		Host       string `env:"POSTGRES_HOST" env-default:"database"`
		Port       int    `env:"POSTGRES_PORT" env-default:"5432"`
		Username   string `env:"POSTGRES_USER" env-default:"postgres"`
		Password   string `env:"POSTGRES_PASSWORD" env-default:"postgres"`
		Name       string `env:"POSTGRES_DB" env-default:"db"`
		AccessType string `yaml:"access-type" env:"DB_ACCESS_TYPE" env-default:"orm"`
	}

	Secrets struct {
		BotToken string `env:"BOT_TOKEN"`
	}
)

type Config struct {
	Env      string   `yaml:"env" env-default:"dev"`
	Serving  Serving  `yaml:"serving"`
	Database Database `yaml:"database"`
	Secrets  Secrets
}

func Load(configFileName string) (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(configFileName, &cfg); err != nil {
		return nil, err
	}

	if cfg.Env == "dev" {
		if err := godotenv.Load(".env"); err != nil {
			return nil, err
		}
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

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
