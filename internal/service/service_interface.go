package service

import (
	"context"

	models "github.com/dimalewshin98-glitch/LTCalendar/internal/model"
)

type ServiceInterface interface {
	Ping(ctx context.Context) error
	Login(ctx context.Context, req models.ApiLoginReq) (models.ApiLoginRes, error)
	Register(ctx context.Context, req models.ApiLoginReq) (models.ApiLoginRes, error)
	AddTest(ctx context.Context, userID int, req models.ApiAddTestReq) (string, error)
	GetTests(ctx context.Context, userID int) (models.ApiGetTestsRes, error)
	GetTest(ctx context.Context, userID int, testUUID string) (models.ApiGetTestRes, error)
	UpdateTest(ctx context.Context, userID int, testUUID string, test models.ApiUpdateTestReq) (string, error)
	Delete(ctx context.Context, userID int, testUUID string) (string, error)
}
