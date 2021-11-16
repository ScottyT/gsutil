package config

import (
	"log"

	"github.com/spf13/viper"
)

// Usage for this is: viperEnvKey("KEY")
func ViperEnvKey(key string) string {
	viper.SetConfigFile("./env/.env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}
	value, ok := viper.Get(key).(string)
	if !ok {
		log.Fatalf("Invalid type assertion")
	}
	return value
}
