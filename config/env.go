package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Usage for this is: viperEnvKey("KEY")
func ViperEnvKey(key string) string {
	viper.SetConfigFile("./.env")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error while reading config file %s", err)
	}
	value, ok := viper.Get(key).(string)
	if !ok {
		fmt.Printf("Invalid type assertion")
	}
	return value
}
