package util

import "github.com/spf13/viper"

// This config struct stores all configuration of the application
// The values are read by viper from a config file or environment variables
type Config struct {
	DbDriver      string `mapstructure:"DB_DRIVER"`
	DbSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
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
