package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	JWTSecret string `yaml:"JWTSecret"`
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

	if conf.JWTSecret == "" {
		return nil, errors.New("JWTSecret is required in the configuration file")
	}

	return &conf, nil
}
