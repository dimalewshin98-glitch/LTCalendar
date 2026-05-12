package repository

import (
	"context"
	"errors"

	models "github.com/dimalewshin98-glitch/LTCalendar/internal/model"
)

var ErrShortURLExists = errors.New("test already exists")
var ErrShortURLDeleted = errors.New("test already deleted")

type RepositoryInterface interface {
	Ping(ctx context.Context) error
	GetUsersID(ctx context.Context) ([]int, error)
	AddTest(ctx context.Context, userID int, req models.ApiAddTestReq, endTime string) (string, error)
	GetTests(ctx context.Context, userID int) (models.ApiGetTestsRes, error)
}
