package config

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

//go:embed meta.json
var replyFS embed.FS

type Config struct {
	Host         string `mapstructure:"HOST"`
	BotPort      string `mapstructure:"BOT_PORT"`
	ScrapperPort string `mapstructure:"SCRAPPER_PORT"`
	BotToken     string `mapstructure:"BOT_TOKEN"`
	Meta         *Meta
}

func MustLoad() (*Config, error) {
	envPath, err := findEnvFile(name)
	if err != nil {
		return nil, err
	}

	viper.SetConfigFile(envPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{}

	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	repliesFile, err := replyFS.ReadFile(repliesName)
	if err != nil {
		return nil, err
	}

	meta := &Meta{}

	if err := json.Unmarshal(repliesFile, meta); err != nil {
		return nil, err
	}

	config.Meta = meta

	return config, nil
}

func findEnvFile(name string) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		target := filepath.Join(currentDir, name)
		if _, err := os.Stat(target); err == nil {
			return target, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}

		currentDir = parentDir
	}

	return "", ErrFailedToFindEnv
}

const (
	name        = ".env"
	repliesName = "meta.json"
)
