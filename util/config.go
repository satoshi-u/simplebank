package util

import (
	"time"

	"github.com/spf13/viper"
)

// This config struct stores all configuration of the application
// The values are read by viper from a config file or environment variables
type Config struct {
	Environment          string        `mapstructure:"ENVIRONMENT"`
	DbDriver             string        `mapstructure:"DB_DRIVER"`
	DbSourceMain         string        `mapstructure:"DB_SOURCE_MAIN"`
	DbSourceTest         string        `mapstructure:"DB_SOURCE_TEST"`
	MigrationUrl         string        `mapstructure:"MIGRATION_URL"`
	HttpServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GrpcServerAddress    string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	ServerType           string        `mapstructure:"SERVER_TYPE"`
	TokenMakerType       string        `mapstructure:"TOKEN_MAKER_TYPE"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
}

// LoadConfig reads configuration from file if path exists or set/override configuration with env-vars if provided
func LoadConfig(path string) (config Config, err error) {
	// read config from file, if exists
	viper.AddConfigPath(path)
	viper.SetConfigName("app") // from app.env
	viper.SetConfigType("env") // from app.env, could also be json, xml, yaml

	// read/override config with env-vars, if exist
	viper.AutomaticEnv()

	// read config now for both cases
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	// unmarshal the config values to target config object
	err = viper.Unmarshal(&config)
	return
}
