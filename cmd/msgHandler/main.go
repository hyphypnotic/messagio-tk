package main

import (
	"github.com/hyphypnotic/messagio-tk/internal/config"
	"github.com/hyphypnotic/messagio-tk/internal/msgHandler/app"
)

func main() {
	cfg, err := config.Load("config/local.yaml")
	if err != nil {
		panic(err)
	}

	msgHandler, err := app.New(cfg)
	if err != nil {
		panic(err)
	}

	msgHandler.Run()
}
