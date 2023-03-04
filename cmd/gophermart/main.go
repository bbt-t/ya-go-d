package main

import (
	"github.com/bbt-t/ya-go-d/internal/app"
	"github.com/bbt-t/ya-go-d/internal/config"
)

func main() {
	/*
		Config generation and start app.
	*/
	cfg := config.NewConfig()
	app.Run(cfg)
}
