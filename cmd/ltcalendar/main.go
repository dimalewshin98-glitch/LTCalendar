package main

import (
	"net/http"

	"github.com/dimalewshin98-glitch/LTCalendar/internal/config"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/handler"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/logger"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/repository"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()
	var repo repository.RepositoryInterface
	var err error
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}
	repo, err = repository.NewDBRepository(cfg.DatabaseDsn)
	logger.Log.Info("Repository adress set to", zap.String("type", cfg.DatabaseDsn))
	if err != nil {
		panic(err)
	}
	app := NewApp(repo, *cfg)
	appHandler := app.GetHandler()
	logger.Log.Info("Running server", zap.String("address", cfg.ServerHostPort))
	err = http.ListenAndServe(cfg.ServerHostPort, logger.RequestLogger(handler.AuthMiddleware(handler.GzipMiddleware(appHandler), repo)))
	if err != nil {
		logger.Log.Fatal("Server failed", zap.Error(err))
		panic(err)
	}
}
