package repository

import (
	"context"
	"errors"

	models "github.com/dimalewshin98-glitch/LTCalendar/internal/model"
)

var ErrTestNotExists = errors.New("test not exists")
var ErrTestDeleted = errors.New("test deleted")
var ErrTestStarted = errors.New("test started")
var ErrDBHostWrongFormat = errors.New("Database host/port wrong format")

type RepositoryInterface interface {
	Ping(ctx context.Context) error
	GetUsersID(ctx context.Context) ([]int, error)
	AddTest(ctx context.Context, userID int, req models.ApiAddTestReq, endTime string) (string, error)
	GetTests(ctx context.Context, userID int, excludeStarted bool) (models.ApiGetTestsRes, error)
	GetTest(ctx context.Context, userID int, testUUID string) (models.ApiGetTestRes, error)
	UpdateTest(ctx context.Context, userID int, testUUID string, test models.ApiUpdateTestReq) (string, error)
	SetTestStarted(ctx context.Context, userID int, testUUID string) (string, error)
	SetDelete(ctx context.Context, userID int, testUUID string) (string, error)
}
