package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Host         string `mapstructure:"HOST"`
	BotPort      string `mapstructure:"BOT_PORT"`
	ScrapperPort string `mapstructure:"SCRAPPER_PORT"`
	BotToken     string `mapstructure:"BOT_TOKEN"`
}

func Load() (*Config, error) {
	envPath, err := findEnvFile(Name)
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
	Name = ".env"
)

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
