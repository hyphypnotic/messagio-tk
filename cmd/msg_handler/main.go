package main

import (
	"github.com/hyphypnotic/messagio-tk/internal/config"
	"github.com/hyphypnotic/messagio-tk/internal/msg_handler/app"
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
