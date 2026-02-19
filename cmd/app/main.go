package main

import (
	"internal/config"

	"github.com/labstack/echo/v4"
)

func main() {
	сfg := config.LoadConfig()
	log := logger.setupLogger(сfg.Env)
	e := echo.New()
	e.HideBanner = true

}
