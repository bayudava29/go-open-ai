package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Configuration struct {
	ENV []Settings `mapstructure:"env"`
}

type Settings struct {
	NAME  string
	VALUE string
}

func InitConfig() {

	viper.SetConfigType("yaml")
	viper.SetConfigName("local.env")
	viper.AddConfigPath("./")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	var env Configuration
	viper.Unmarshal(&env)
	for _, value := range env.ENV {
		if _, ok := os.LookupEnv(value.NAME); !ok {
			os.Setenv(value.NAME, value.VALUE)
		}

	}
}
