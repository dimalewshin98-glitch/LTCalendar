package main

import (
	"net/http"

	"github.com/dimalewshin98-glitch/LTCalendar/internal/config"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/handler"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/repository"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/service"
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
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", requestsHandler.Ping)
	api := http.NewServeMux()
	api.HandleFunc("POST /tests", requestsHandler.AddTest)
	api.HandleFunc("GET /tests", requestsHandler.GetTests)
	api.HandleFunc("GET /tests/{uuid}", requestsHandler.GetTest)
	api.HandleFunc("PUT /tests/{uuid}", requestsHandler.UpdateTest)
	api.HandleFunc("DELETE /tests/{uuid}", requestsHandler.DeleteTest)
	api.HandleFunc("POST /user/login", requestsHandler.LoginUser)
	api.HandleFunc("POST /user/register", requestsHandler.RegisterUser)
	mux.Handle("/api/", http.StripPrefix("/api", api))
	return mux
}
