package main

import (
	"internal/config"
	errorhandler "main/pkg/errDTO/error_handler"

	"github.com/go-playground/validator/v10"

	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	сfg := config.LoadConfig()
	log := logger.setupLogger(сfg.Env)
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	e.HideBanner = true
	e.HTTPErrorHandler = errorhandler.CustomHTTPErrorHandler

}
