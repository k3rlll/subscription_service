package main

import (
	"context"
	"fmt"
	"main/internal/app"
	"main/internal/config"
	conn "main/internal/database/postgres"
	repo "main/internal/database/postgres/repository"
	"main/internal/usecase"
	"main/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	logger := logger.SetupLogger(cfg.Env)

	//database connection
	pool, err := conn.NewPostgresConnection(cfg.Database.DSN())
	if err != nil {
		logger.Error("Failed to connect to the database", "error", err.Error())
		panic(err)
	}
	defer pool.Close()

	// setup repository and usecase
	repo := repo.NewRepository(pool)
	usecase := usecase.NewUsecase(repo)

	e := app.Run(cfg, logger, usecase)

	go func() {
		if err := e.Start(":" + fmt.Sprintf("%d", cfg.Server.Port)); err != nil && err != http.ErrServerClosed {
			logger.Error("Server forced to shutdown", "error", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down server gracefully...")

	// 10 seconds timeout before force shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// echo shutdown
	if err := e.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", "error", err.Error())
	}

	logger.Info("Server exited properly")
}
