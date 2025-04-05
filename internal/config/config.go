package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Secrets Secrets
	Env     string  `yaml:"env" env-default:"local"`
	Serving Serving `yaml:"serving"`
}

type Serving struct {
	Host         string `yaml:"host" env-default:"localhost"`
	ScrapperPort string `yaml:"scrapper_port" env-default:"8080"`
	BotPort      string `yaml:"bot_port" env-default:"8081"`
}

type Secrets struct {
	BotToken string `env:"BOT_TOKEN"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig("config/config.yaml", &cfg); err != nil {
		return nil, err
	}

	if cfg.Env == "local" {
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
		Description: "registrates the user",
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
