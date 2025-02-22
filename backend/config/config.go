package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	DatabaseUrl string `mapstructure:"DATABASE_URL"`
	ExternalApi string `mapstructure:"EXTERNAL_API"`
}

func validateConfig(config *Config) error {
	configMap := map[string]interface{}{
		"DATABASE_URL": config.DatabaseUrl,
		"EXTERNAL_API": config.ExternalApi,
	}

	for key, value := range configMap {
		if isEmptyValue(value) {
			return fmt.Errorf("missing required configuration field: %s", key)
		}
	}

	return nil
}

func LoadConfig(path string) (config *Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("../.env")
	viper.SetConfigType("env")

	// Automatically map environment variables
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	err = validateConfig(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func Load(path string) error {
	err := godotenv.Load(path)
	if err != nil {
		return err
	}

	return nil
}

func isEmptyValue(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case int64:
		return v == 0
	case nil:
		return true
	default:
		return false
	}
}
