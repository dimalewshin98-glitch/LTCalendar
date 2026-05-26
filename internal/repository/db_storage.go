package repository

import (
	"context"
	"database/sql"
	_ "embed"
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

//go:embed migrations/000002_create_users_table.up.sql
var sqlCreateUsersTable string

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
		return nil, ErrDBHostWrongFormat
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
	if !slices.Contains(tables, "users") {
		_, err = tx.ExecContext(ctx, sqlCreateUsersTable)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return rbErr
			}
			return err
		}
	}
	return tx.Commit()
}

func (r *DBRepository) Login(ctx context.Context, req models.ApiLoginReq) (int, string, error) {
	sqlSelect := "SELECT user_id, user_password FROM users WHERE user_login = $1"
	row := r.dbConnection.QueryRowContext(ctx, sqlSelect, req.Login)
	var userID int
	var userPass string
	err := row.Scan(&userID, &userPass)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", ErrUserLogPassIncorrect
		}
		return 0, "", err
	}
	return userID, userPass, nil
}

func (r *DBRepository) Register(ctx context.Context, req models.ApiLoginReq) (int, error) {
	tx, err := r.dbConnection.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	sqlSelect := "SELECT MAX(user_id) FROM users;"
	row := r.dbConnection.QueryRowContext(ctx, sqlSelect)
	var maxUserId int
	err = row.Scan(&maxUserId)
	if err != nil {
		if err.Error() == ErrColumnIndexZero.Error() {
			maxUserId = 0
		} else {
			return 0, err
		}
	}
	newUserId := maxUserId + 1
	sqlInsert := "INSERT INTO users (user_id, user_login, user_password) VALUES ($1, $2, $3) ON CONFLICT (user_login) DO NOTHING RETURNING user_id;"
	var userID string
	row = tx.QueryRowContext(ctx, sqlInsert, newUserId, req.Login, req.Password)
	err = row.Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrUserAlreadyExists
		} else {
			rbErr := tx.Rollback()
			if rbErr != nil {
				return 0, rbErr
			} else {
				return 0, err
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return newUserId, err
}

func (r *DBRepository) GetUsersID(ctx context.Context) ([]int, error) {
	sqlSelect := "SELECT DISTINCT user_id FROM users WHERE user_id IS NOT NULL;"
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
		sqlInsert := "INSERT INTO tests (uuid, user_id, test_name, tps, start_time, end_time, additional_params, is_deleted, is_started) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);"
		tx.QueryRowContext(ctx, sqlInsert, UUID, userID, req.TestName, req.TPS, req.StartTime, endTime, req.AdditionalParams, false, false)
		err := tx.Commit()
		if err != nil {
			return "", err
		}
		return UUID, err
	}
}

func (r *DBRepository) GetTests(ctx context.Context, userID int, excludeStarted bool) (models.ApiGetTestsRes, error) {
	var sqlSelect string
	sqlSelect = "SELECT uuid, test_name, start_time, is_started FROM tests where is_deleted = false"
	if excludeStarted {
		sqlSelect += " and is_started=false"
	}
	var rows *sql.Rows
	var err error
	if userID != 0 {
		sqlSelect += " and user_id = $1"
		rows, err = r.dbConnection.QueryContext(ctx, sqlSelect, userID)
	} else {
		rows, err = r.dbConnection.QueryContext(ctx, sqlSelect)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var userTests models.ApiGetTestsRes
	for rows.Next() {
		var userTest models.TestRes
		err := rows.Scan(&userTest.TestUID, &userTest.TestName, &userTest.StartTime, &userTest.IsStarted)
		if err != nil {
			return nil, err
		}
		userTests = append(userTests, userTest)
	}
	return userTests, nil
}

func (r *DBRepository) GetTest(ctx context.Context, userID int, testUUID string) (models.ApiGetTestRes, error) {
	sqlSelect := "SELECT test_name, start_time, end_time, tps, additional_params, is_started, is_deleted FROM tests where uuid = $1"
	var row *sql.Row
	if userID != 0 {
		sqlSelect += " and user_id = $2"
		row = r.dbConnection.QueryRowContext(ctx, sqlSelect, testUUID, userID)
	} else {
		row = r.dbConnection.QueryRowContext(ctx, sqlSelect, testUUID)
	}
	var userTest models.ApiGetTestRes
	var isDeleted bool
	err := row.Scan(&userTest.TestName, &userTest.StartTime, &userTest.EndTime, &userTest.TPS, &userTest.AdditionalParams, &userTest.IsStarted, &isDeleted)
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
		return testUUID, err
	}
	defer tx.Rollback()
	sqlInsert := `
        UPDATE tests
        SET test_name = $1,
            start_time = $2,
            end_time = $3,
            tps = $4,
            additional_params = $5
        WHERE user_id = $6 AND uuid = $7 AND is_deleted = false RETURNING uuid;
	`
	var UUID string
	row := tx.QueryRowContext(ctx, sqlInsert, test.TestName, test.StartTime, test.EndTime, test.TPS, test.AdditionalParams, userID, testUUID)
	err = row.Scan(&UUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return testUUID, ErrTestDeleted
		}
		return testUUID, err
	}
	err = tx.Commit()
	if err != nil {
		return testUUID, err
	}
	return testUUID, nil
}

func (r *DBRepository) SetTestStarted(ctx context.Context, userID int, testUUID string) (string, error) {
	tx, err := r.dbConnection.BeginTx(ctx, nil)
	if err != nil {
		return testUUID, err
	}
	defer tx.Rollback()
	sqlInsert := `
        UPDATE tests
        SET is_started = true
        WHERE uuid = $1 AND is_deleted = false
	`
	var row *sql.Row
	if userID != 0 {
		sqlInsert += " and user_id = $2 RETURNING uuid"
		row = r.dbConnection.QueryRowContext(ctx, sqlInsert, testUUID, userID)
	} else {
		sqlInsert += " RETURNING uuid"
		row = r.dbConnection.QueryRowContext(ctx, sqlInsert, testUUID)
	}
	var UUID string
	err = row.Scan(&UUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return testUUID, ErrTestStarted
		}
		return testUUID, err
	}
	err = tx.Commit()
	if err != nil {
		return testUUID, err
	}
	return testUUID, nil
}

func (r *DBRepository) SetDelete(ctx context.Context, userID int, testUUID string) (string, error) {
	var ctxUUID string
	var singleReq bool
	var isLastReq bool
	var err error
	var tx *sql.Tx
	if ctx.Value("UUID") == nil {
		ctxUUID = ""
		singleReq = true
		isLastReq = true
	} else {
		ctxUUID = ctx.Value("UUID").(string)
		singleReq = false
		if ctx.Value("isLastReq") == nil {
			isLastReq = false
		} else {
			isLastReq = ctx.Value("isLastReq").(bool)
		}
	}
	if singleReq {
		tx = nil
	} else {
		tx = r.txMap[ctxUUID]
	}
	if tx == nil {
		tx, err = r.dbConnection.BeginTx(ctx, nil)
		if err != nil {
			return testUUID, err
		}
	}
	if !singleReq {
		r.txMap[ctxUUID] = tx
	}
	select {
	case <-ctx.Done():
		rbErr := tx.Rollback()
		if !singleReq {
			delete(r.txMap, ctxUUID)
		}
		return testUUID, rbErr
	default:
		sqlInsert := "UPDATE tests SET is_deleted = true WHERE uuid = $1 and user_id = $2 RETURNING uuid;"
		var UUID string
		row := tx.QueryRowContext(ctx, sqlInsert, testUUID, userID)
		err = row.Scan(&UUID)
		if err != nil {
			if err != sql.ErrNoRows {
				if rbErr := tx.Rollback(); rbErr != nil {
					return testUUID, rbErr
				}
				return testUUID, err
			}
		}
		if isLastReq {
			commitErr := tx.Commit()
			if !singleReq {
				delete(r.txMap, ctxUUID)
			}
			if commitErr == nil {
				return testUUID, err
			} else {
				return testUUID, commitErr
			}
		}
		return testUUID, err
	}
}

func (r *DBRepository) Ping(ctx context.Context) error {
	err := r.dbConnection.PingContext(ctx)
	return err
}
