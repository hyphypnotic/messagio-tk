package main

import (
	"github.com/hyphypnotic/messagio-tk/internal/config"
	"github.com/hyphypnotic/messagio-tk/internal/orchestrator/app"
)

func main() {
	cfg, err := config.Load("config/local.yaml")
	if err != nil {
		panic(err)
	}

	orchestrator, err := app.New(cfg)
	if err != nil {
		panic(err)
	}

	err = orchestrator.Run()
	if err != nil {
		panic(err)
	}
}
