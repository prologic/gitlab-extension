package config

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
)

// Configuration file type.
type Config struct {
	Port             int      `yaml:"port"`
	GitlabUri        string   `yaml:"gitlab-uri"`
	GitlabToken      string   `yaml:"gitlab-token"`
	BotToken         string   `yaml:"telegram-bot-token"`
	GitlabNamespaces []string `yaml:"gitlab-namespaces"`
	Origins          []string `yaml:"origins"`
}

// Loads config file.
// filepath - path to config file.
func Get(filepath string, logger *logrus.Logger) *Config {
	c := new(Config)
	file, err := os.Open(filepath)
	if err != nil {
		logger.Fatalf("Config load err: %v", err)
	}
	defer file.Close()
	err = yaml.NewDecoder(file).Decode(c)
	if err != nil {
		logger.Fatalf("Config unmarshal err: %v", err)
	}
	return c
}
