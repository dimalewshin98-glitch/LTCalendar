package service

import (
	"context"

	"github.com/dimalewshin98-glitch/LTCalendar/internal/config"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/repository"
)

type ShorterService struct {
	repo   repository.RepositoryInterface
	config *config.Config
}

func NewShorterService(repo repository.RepositoryInterface, config *config.Config) *ShorterService {
	serviceInstance := &ShorterService{
		repo:   repo,
		config: config,
	}
	return serviceInstance
}

func (s *ShorterService) Ping(ctx context.Context) error {
	err := s.repo.Ping(ctx)
	return err
}
