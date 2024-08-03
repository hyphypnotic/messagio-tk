package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

// Config - configuration settings
type Config struct {
	Env      string `yaml:"env" env-default:"local"`
	Postgres struct {
		Host     string `yaml:"host" env-default:"localhost"`
		Port     int    `yaml:"port" env-default:"5432"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"database"`
	Kafka struct {
		Brokers  []string `yaml:"brokers"`
		Group    string   `yaml:"group"`
		Topic    string   `yaml:"topic"`
		Consumer struct {
			RebalanceStrategy string `yaml:"rebalance_strategy"`
			OffsetInitial     string `yaml:"offset_initial"`
		} `yaml:"consumer"`
	} `yaml:"kafka"`
	GRPC struct {
		Port    int    `yaml:"port"`
		Timeout int    `yaml:"timeout"`
		Host    string `yaml:"host"`
	} `yaml:"grpc"`

	HttpPort int `yaml:"http_address"`
}

// MustLoad loads config from a .yaml file
func MustLoad(filePath string) (*Config, error) {
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
