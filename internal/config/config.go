package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

// Config - configuration settings
type Config struct {
	Env      string `yaml:"env"`
	Postgres struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"postgres"`
	Kafka struct {
		Brokers []string            `yaml:"brokers"`
		Topics  map[string][]string `yaml:"topics"`
	} `yaml:"kafka"`

	GRPC struct {
		Port    int    `yaml:"port"`
		Timeout int    `yaml:"timeout"`
		Host    string `yaml:"host"`
	} `yaml:"grpc"`

	HttpPort int `yaml:"http_port"`
}

// Load loads config from a .yaml file
func Load(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
