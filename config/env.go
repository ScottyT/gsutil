package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	WebAppUrl      string `mapstructure:"WEB_APP_URL"`
	StorageBucket  string `mapstructure:"STORAGE_BUCKET"`
	CredentialFile string `mapstructure:"CREDENTIAL_FILE"`
	ProjectId      string `mapstructure:"PROJECT_ID"`
	SaEmail        string `mapstructure:"SERVICE_ACCOUNT_EMAIL"`
}

// Usage for this is: viperEnvKey("KEY")
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error while reading config file %s", err)
		return
	}
	err = viper.Unmarshal(&config)
	return
}
