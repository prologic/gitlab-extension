package conf

import (
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var availableLoggingTargets = [2]string{"console", "file"}

const (
	defaultLogFilePath = "logs/gitlab-ext.log"
	defaultMaxSizeMb   = 30
	defaultMaxFiles    = 3
	defaultMaxAgeDays  = 14
)

// Configuration file type.
type Config struct {
	Port                int                  `yaml:"port"`
	GitlabUri           string               `yaml:"gitlab-uri"`
	GitlabToken         string               `yaml:"gitlab-token"`
	BotEnabled          bool                 `yaml:"telegram-bot-enabled"`
	BotToken            string               `yaml:"telegram-bot-token"`
	GitlabNamespaces    []string             `yaml:"gitlab-namespaces"`
	LogTo               []string             `yaml:"log-to"`
	RollingFileSettings *RollingFileSettings `yaml:"rolling-file-settings,omitempty"`
}

// Settings for rolling file logging.
// Ignored if 'file' doesn't exist in Config.LogTo array.
type RollingFileSettings struct {
	FileName   string `yaml:"file-name"`
	MaxSizeMb  int    `yaml:"max-size-mb"`
	MaxFiles   int    `yaml:"max-files"`
	MaxAgeDays int    `yaml:"max-age-days"`
	Compress   bool   `yaml:"compress"`
}

// True if has file logging target.
func (c *Config) HasFileLogging() bool {
	return funk.Contains(c.LogTo, "file")
}

// True if has console logging target.
func (c *Config) HasConsoleLogging() bool {
	return funk.Contains(c.LogTo, "console")
}

// Creates lumberjack logger from config.
func (s *RollingFileSettings) CreateLumberjack(logger *logrus.Logger) *lumberjack.Logger {
	result := lumberjack.Logger{}
	if s.FileName == "" {
		result.Filename = defaultLogFilePath
		logger.Warnf("Empty log filepath. Setting default value: %s", defaultLogFilePath)
	} else {
		result.Filename = s.FileName
	}

	if s.MaxSizeMb <= 0 {
		result.MaxSize = defaultMaxSizeMb
		logger.Warnf("Invalid max size %d. Setting default value: %d", s.MaxSizeMb, defaultMaxSizeMb)
	} else {
		result.MaxSize = s.MaxSizeMb
	}

	if s.MaxFiles <= 0 {
		result.MaxBackups = defaultMaxFiles
		logger.Warnf("Invalid max files count %d. Setting default value: %d", s.MaxFiles, defaultMaxFiles)
	} else {
		result.MaxBackups = s.MaxFiles
	}

	if s.MaxAgeDays <= 0 {
		result.MaxAge = defaultMaxAgeDays
		logger.Warnf("Invalid max age %d. Setting default value: %d", s.MaxAgeDays, defaultMaxAgeDays)
	} else {
		result.MaxAge = s.MaxAgeDays
	}

	result.Compress = s.Compress

	return &result
}

// Loads config file.
// filepath - path to config file.
func Get(filepath string, logger *logrus.Logger) *Config {
	var c Config
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		logger.Fatalf("Config load err: %v", err)
	}
	err = yaml.Unmarshal(file, &c)
	if err != nil {
		logger.Fatalf("Config unmarshal err: %v", err)
	}
	if len(c.LogTo) > len(availableLoggingTargets) {
		logger.Fatal("To many values in 'Config.LogTo'")
	}
	for _, target := range c.LogTo {
		if !funk.Contains(availableLoggingTargets, target) {
			logger.Fatalf("Unrecognized logging target in 'Config.LogTo': %v", target)
		}
	}
	if !c.HasFileLogging() {
		c.RollingFileSettings = nil
	}

	return &c
}
