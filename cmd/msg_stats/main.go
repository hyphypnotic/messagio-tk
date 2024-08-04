package main

import (
	"github.com/hyphypnotic/messagio-tk/internal/config"
	"github.com/hyphypnotic/messagio-tk/internal/msg_stats/app"
)

func main() {
	cfg, err := config.Load("config/local.yaml")
	if err != nil {
		panic(err)
	}

	msgStats, err := app.New(cfg)
	if err != nil {
		panic(err)
	}

	err = msgStats.Run()
	if err != nil {
		panic(err)
	}
}
