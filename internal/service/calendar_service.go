package service

import (
	"context"
	"errors"
	"time"

	"github.com/dimalewshin98-glitch/LTCalendar/internal/config"
	models "github.com/dimalewshin98-glitch/LTCalendar/internal/model"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/repository"
)

var ErrParsingDate = errors.New("error parsing date. exmaple: 2006-01-02T15:04:05+03:00")

type CalendarService struct {
	repo   repository.RepositoryInterface
	config *config.Config
}

func NewCalendarService(repo repository.RepositoryInterface, config *config.Config) *CalendarService {
	serviceInstance := &CalendarService{
		repo:   repo,
		config: config,
	}
	return serviceInstance
}

func (s *CalendarService) AddTest(ctx context.Context, userID int, req models.ApiAddTestReq) (string, error) {
	t, err := parseTime(req.StartTime)
	if err != nil {
		return "", err
	}
	endTime := countEndTime(t, req.DurationMin)
	testUUID, err := s.repo.AddTest(ctx, userID, req, endTime)
	return testUUID, err
}

func (s *CalendarService) GetTests(ctx context.Context, userID int) (models.ApiGetTestsRes, error) {
	tests, err := s.repo.GetTests(ctx, userID)
	return tests, err
}

func (s *CalendarService) GetTest(ctx context.Context, userID int, testUUID string) (models.ApiGetTestRes, error) {
	test, err := s.repo.GetTest(ctx, userID, testUUID)
	return test, err
}

func (s *CalendarService) UpdateTest(ctx context.Context, userID int, testUUID string, test models.ApiUpdateTestReq) (string, error) {
	testUUID, err := s.repo.UpdateTest(ctx, userID, testUUID, test)
	return testUUID, err
}

func (s *CalendarService) Ping(ctx context.Context) error {
	err := s.repo.Ping(ctx)
	return err
}

func parseTime(value string) (time.Time, error) {
	layout := "2006-01-02T15:04:05+03:00"
	t, err := time.Parse(layout, value)
	if err != nil {
		return time.Now(), ErrParsingDate
	}
	return t, nil
}

func countEndTime(startTime time.Time, duration int) string {
	layout := "2006-01-02T15:04:05+03:00"
	endTime := startTime.Add(time.Duration(duration) * time.Minute)
	endTimeStr := endTime.Format(layout)
	return endTimeStr
}
