package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Addr string `yaml:"address" env:"ADDRESS" env-default:"localhost:8082"`
}
type Config struct {
	Env         string `yaml:"env" env:"ENV" env-required:"true" env-default:"PROD"`
	URL         string `yaml:"URL" env:"URL" env-required:"true"`
	DBName      string `yaml:"DB_NAME" env:"DB_NAME" env-required:"true"`
	HTTPServer  `yaml:"http_server" env-required:"true"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		flags := flag.String("config", "", "Path of Config file ")
		flag.Parse()

		configPath = *flags

	}

	if configPath == "" {
		log.Fatal("Config path is not set")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file is not exist: %s", configPath)
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("Cannot read Config file : %s", err.Error())
	}

	return &cfg
}
