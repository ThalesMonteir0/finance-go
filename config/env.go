package config

import (
	"strings"

	"github.com/spf13/viper"
)

var Env *EnvConfig

type EnvConfig struct {
	VerbooAPIKey string
}

func LoadEnv() *EnvConfig {
	viper.SetDefault("VERBOO_API_KEY", "token-default")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	_ = viper.ReadInConfig() // ignora erro: arquivo .env é opcional

	Env = &EnvConfig{
		VerbooAPIKey: viper.GetString("VERBOO_API_KEY"),
	}

	return Env
}
