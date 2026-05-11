package repository

import (
	"context"
	"errors"
)

var ErrShortURLExists = errors.New("test already exists")
var ErrShortURLDeleted = errors.New("test already deleted")

type RepositoryInterface interface {
	Ping(ctx context.Context) error
	GetUsersID(ctx context.Context) ([]int, error)
}
