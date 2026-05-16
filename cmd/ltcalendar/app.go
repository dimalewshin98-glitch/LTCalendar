package main

import (
	"net/http"

	"github.com/dimalewshin98-glitch/LTCalendar/internal/config"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/handler"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/repository"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/service"
	"github.com/go-chi/chi/v5"
)

type App struct {
	repo repository.RepositoryInterface
	cfg  config.Config
}

func NewApp(repo repository.RepositoryInterface, cfg config.Config) *App {
	return &App{
		repo: repo,
		cfg:  cfg,
	}
}

func (a *App) GetHandler() http.Handler {
	shorterService := service.NewCalendarService(a.repo, &a.cfg)
	requestsHandler := handler.NewRequestsHandler(shorterService)
	r := chi.NewRouter()
	r.Get("/ping", requestsHandler.Ping)
	r.Post("/api/tests", requestsHandler.AddTest)
	r.Get("/api/tests", requestsHandler.GetTests)
	r.Get("/api/tests/{uuid}", requestsHandler.GetTest)
	r.Put("/api/tests/{uuid}", requestsHandler.UpdateTest)
	r.Delete("/api/tests/{uuid}", requestsHandler.DeleteTest)
	return r
}
