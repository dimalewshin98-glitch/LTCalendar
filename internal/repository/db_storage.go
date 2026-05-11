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

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/000001_create_urls_table.up.sql
var sqlCreateUrlsTable string

//go:embed migrations/000002_add_unique_index_to_orig_url.up.sql
var sqlAddUniqueIndexToOrigUrl string

//go:embed migrations/000003_add_user_id_column.up.sql
var sqlAddUserIdColumn string

//go:embed migrations/000004_add_deleted_flag_column.up.sql
var sqlAddDeletedFlagColumn string

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
	if !slices.Contains(tables, "urls") {
		_, err = tx.ExecContext(ctx, sqlCreateUrlsTable)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return rbErr
			}
			return err
		}
		_, err = tx.ExecContext(ctx, sqlAddUniqueIndexToOrigUrl)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return rbErr
			}
			return err
		}
		_, err = tx.ExecContext(ctx, sqlAddUserIdColumn)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return rbErr
			}
			return err
		}
		_, err = tx.ExecContext(ctx, sqlAddDeletedFlagColumn)
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
	sqlSelect := "SELECT DISTINCT user_id FROM urls WHERE user_id IS NOT NULL;"
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

func (r *DBRepository) Ping(ctx context.Context) error {
	err := r.dbConnection.PingContext(ctx)
	return err
}
