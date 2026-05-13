package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	models "github.com/dimalewshin98-glitch/LTCalendar/internal/model"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/000001_create_tests_table.up.sql
var sqlCreateTestsTable string

type DBRepository struct {
	dbDsn        string
	dbConnection *sql.DB
	txMap        map[string]*sql.Tx
}

func NewDBRepository(dbDsn string) (*DBRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	hostPort := strings.Split(dbDsn, ":")
	var ps string
	if len(hostPort) == 1 {
		ps = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
			hostPort[0], `postgres`, `admin`, `postgres`)
	} else if len(hostPort) == 2 {
		ps = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			hostPort[0], hostPort[1], `postgres`, `admin`, `postgres`)
	} else {
		return nil, errors.New("Database host/port wrong format")
	}
	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}
	dbRepository := &DBRepository{
		dbDsn:        dbDsn,
		dbConnection: db,
		txMap:        make(map[string]*sql.Tx)}
	err = dbRepository.Ping(ctx)
	if err != nil {
		return nil, err
	}
	err = dbRepository.CreateTables(ctx)
	if err != nil {
		return nil, err
	}
	return dbRepository, nil
}

func (r *DBRepository) CreateTables(ctx context.Context) error {
	tx, err := r.dbConnection.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	sqlReq := "SELECT table_name FROM information_schema.tables WHERE table_schema='public';"
	rows, err := tx.QueryContext(ctx, string(sqlReq))
	if err != nil {
		return nil
	}
	var tableName string
	var tables []string
	for rows.Next() {
		rows.Scan(&tableName)
		tables = append(tables, tableName)
	}
	if !slices.Contains(tables, "tests") {
		_, err = tx.ExecContext(ctx, sqlCreateTestsTable)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return rbErr
			}
			return err
		}
	}
	return tx.Commit()
}

func (r *DBRepository) GetUsersID(ctx context.Context) ([]int, error) {
	sqlSelect := "SELECT DISTINCT user_id FROM tests WHERE user_id IS NOT NULL;"
	rows, err := r.dbConnection.QueryContext(ctx, sqlSelect)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var usersID []int
	for rows.Next() {
		var userID int
		err := rows.Scan(&userID)
		if err != nil {
			return nil, err
		}
		usersID = append(usersID, userID)
	}
	return usersID, nil
}

func (r *DBRepository) AddTest(ctx context.Context, userID int, req models.ApiAddTestReq, endTime string) (string, error) {
	UUID := uuid.New().String()
	tx, err := r.dbConnection.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	select {
	case <-ctx.Done():
		rbErr := tx.Rollback()
		return "", rbErr
	default:
		sqlInsert := "INSERT INTO tests (uuid, user_id, test_name, tps, start_time, end_time, additional_params, is_deleted) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
		tx.QueryRowContext(ctx, sqlInsert, UUID, userID, req.TestName, req.TPS, req.StartTime, endTime, req.AdditionalParams, false)
		err := tx.Commit()
		if err != nil {
			return "", err
		}
		return UUID, err
	}
}

func (r *DBRepository) GetTests(ctx context.Context, userID int) (models.ApiGetTestsRes, error) {
	sqlSelect := "SELECT uuid, test_name FROM tests where user_id = $1;"
	rows, err := r.dbConnection.QueryContext(ctx, sqlSelect, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var userTests models.ApiGetTestsRes
	for rows.Next() {
		var userTest models.TestRes
		err := rows.Scan(&userTest.TestUID, &userTest.TestName)
		if err != nil {
			return nil, err
		}
		userTests = append(userTests, userTest)
	}
	return userTests, nil
}

func (r *DBRepository) GetTest(ctx context.Context, userID int, testUUID string) (models.ApiGetTestRes, error) {
	sqlSelect := "SELECT test_name, start_time, end_time, tps, additional_params, is_deleted FROM tests where user_id = $1 and uuid = $2;"
	row := r.dbConnection.QueryRowContext(ctx, sqlSelect, userID, testUUID)
	var userTest models.ApiGetTestRes
	var isDeleted bool
	err := row.Scan(&userTest.TestName, &userTest.StartTime, &userTest.EndTime, &userTest.TPS, &userTest.AdditionalParams, &isDeleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.ApiGetTestRes{}, ErrTestNotExists
		}
		return models.ApiGetTestRes{}, err
	}
	if isDeleted {
		return models.ApiGetTestRes{}, ErrTestDeleted
	}
	return userTest, nil
}

func (r *DBRepository) UpdateTest(ctx context.Context, userID int, testUUID string, test models.ApiUpdateTestReq) (string, error) {
	tx, err := r.dbConnection.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	sqlInsert := `
        UPDATE tests
        SET test_name = $1,
            start_time = $2,
            end_time = $3,
            tps = $4,
            additional_params = $5
        WHERE user_id = $6 AND uuid = $7
	`
	_ = tx.QueryRowContext(ctx, sqlInsert, test.TestName, test.StartTime, test.EndTime, test.TPS, test.AdditionalParams, userID, testUUID)
	err = tx.Commit()
	if err != nil {
		return "", err
	}
	return testUUID, nil
}

func (r *DBRepository) Ping(ctx context.Context) error {
	err := r.dbConnection.PingContext(ctx)
	return err
}
