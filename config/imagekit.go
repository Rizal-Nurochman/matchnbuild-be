package config

import (
	"github.com/spf13/viper"
)

type ImageKitConfig struct {
	PublicKey   string `mapstructure:"IMAGEKIT_PUBLIC_KEY"`
	PrivateKey  string `mapstructure:"IMAGEKIT_PRIVATE_KEY"`
	URLEndpoint string `mapstructure:"IMAGEKIT_URL_ENDPOINT"`
}

func NewImageKitConfig() (*ImageKitConfig, error) {
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	viper.AutomaticEnv()

	var config ImageKitConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
