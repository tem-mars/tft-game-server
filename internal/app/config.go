package app

import (
    "github.com/spf13/viper"
)

type Config struct {
    Server struct {
        Port string
        Host string
    }
    Database struct {
        Host     string
        Port     string
        User     string
        Password string
        DBName   string
    }
}

func LoadConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AutomaticEnv()

    var config Config

    // Default values
    viper.SetDefault("server.port", "8080")
    viper.SetDefault("server.host", "0.0.0.0")

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    return &config, nil
}