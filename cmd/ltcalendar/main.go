package main

import (
	"net/http"
	"strconv"

	"github.com/dimalewshin98-glitch/LTCalendar/internal/config"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/handler"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/logger"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/repository"
)

func main() {
	cfg := config.NewConfig()
	var repo repository.RepositoryInterface
	var err error
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}
	repo, err = repository.NewDBRepository(cfg.DatabaseDsn)
	logger.Log.Info("Start tests in thread", "enabled", strconv.FormatBool(cfg.EnabledTestStarter))
	logger.Log.Info("Repository adress set to", "type", cfg.DatabaseDsn)
	if err != nil {
		panic(err)
	}
	app := NewApp(repo, *cfg)
	appHandler := app.GetHandler()
	logger.Log.Info("Running server", "address", cfg.ServerHostPort)
	err = http.ListenAndServe(cfg.ServerHostPort, logger.RequestLogger(handler.AuthMiddleware(handler.GzipMiddleware(appHandler), repo, cfg.SecretKey)))
	if err != nil {
		logger.Log.Error("Server failed", "error", err)
		panic(err)
	}
}
