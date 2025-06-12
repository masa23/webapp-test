package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AccessToken struct {
		JWTSecret string        `yaml:"JWTSecret"`
		Duration  time.Duration `yaml:"Duration"`
	} `yaml:"AccessToken"`
	RefreshToken struct {
		Duration time.Duration `yaml:"Duration"`
	} `yaml:"RefreshToken"`
}

func Load(path string) (*Config, error) {
	var conf Config
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(buf, &conf); err != nil {
		return nil, err
	}

	if conf.AccessToken.JWTSecret == "" {
		return nil, errors.New("JWTSecret is required in the configuration file")
	}

	if conf.AccessToken.Duration < 1 {
		conf.AccessToken.Duration = time.Minute
	}

	if conf.RefreshToken.Duration < 1 {
		conf.RefreshToken.Duration = time.Hour * 24 * 7 // Default 7d
	}

	return &conf, nil
}
