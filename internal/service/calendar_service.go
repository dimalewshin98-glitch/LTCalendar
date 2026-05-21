package service

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/dimalewshin98-glitch/LTCalendar/internal/config"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/logger"
	models "github.com/dimalewshin98-glitch/LTCalendar/internal/model"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const ENCRYPT_KEY_B64 = "CCHki0M3gumr8P8fs6I7IkyHHdUeNpEbV1/lpbQOtGg="

var ErrParsingDate = errors.New("Error parsing date. Need date format: YYYY-MM-DDTHH:MM:SS±HH:MM")
var ErrCompareDate = errors.New("End time must be after start time")

type CalendarService struct {
	repo    repository.RepositoryInterface
	config  *config.Config
	msgChan chan models.TestDeleteMessage
}

func NewCalendarService(repo repository.RepositoryInterface, config *config.Config) *CalendarService {
	serviceInstance := &CalendarService{
		repo:    repo,
		config:  config,
		msgChan: make(chan models.TestDeleteMessage, 1024),
	}
	go serviceInstance.flushMessages()
	if config.EnabledTestStarter {
		go serviceInstance.scheduleTests()
	}
	return serviceInstance
}

func (s *CalendarService) Login(ctx context.Context, req models.ApiLoginReq) (models.ApiLoginRes, error) {
	userID, userEncryptedPass, err := s.repo.Login(ctx, req)
	if err != nil {
		return models.ApiLoginRes{}, err
	}
	userDecryptedPass, err := s.decryptPass(userEncryptedPass)
	if err != nil {
		return models.ApiLoginRes{}, repository.ErrUserLogPassIncorrect
	}
	if userDecryptedPass != req.Password {
		return models.ApiLoginRes{}, repository.ErrUserLogPassIncorrect
	}
	res := models.ApiLoginRes{UserID: userID}
	return res, err
}

func (s *CalendarService) Register(ctx context.Context, req models.ApiLoginReq) (models.ApiLoginRes, error) {
	encryptedPass, err := s.encryptPass(req.Password)
	if err != nil {
		return models.ApiLoginRes{}, repository.ErrUserLogPassIncorrect
	}
	req.Password = encryptedPass
	userID, err := s.repo.Register(ctx, req)
	if err != nil {
		return models.ApiLoginRes{}, err
	}
	res := models.ApiLoginRes{UserID: userID}
	return res, err
}

func (s *CalendarService) AddTest(ctx context.Context, userID int, req models.ApiAddTestReq) (string, error) {
	t, err := s.parseTime(req.StartTime)
	if err != nil {
		return "", err
	}
	endTime := s.countEndTime(t, req.DurationMin)
	testUUID, err := s.repo.AddTest(ctx, userID, req, endTime)
	return testUUID, err
}

func (s *CalendarService) GetTests(ctx context.Context, userID int) (models.ApiGetTestsRes, error) {
	tests, err := s.repo.GetTests(ctx, userID, false)
	return tests, err
}

func (s *CalendarService) GetTest(ctx context.Context, userID int, testUUID string) (models.ApiGetTestRes, error) {
	test, err := s.repo.GetTest(ctx, userID, testUUID)
	return test, err
}

func (s *CalendarService) UpdateTest(ctx context.Context, userID int, testUUID string, test models.ApiUpdateTestReq) (string, error) {
	startTime, err := s.parseTime(test.StartTime)
	if err != nil {
		return testUUID, err
	}
	endTime, err := s.parseTime(test.EndTime)
	if err != nil {
		return testUUID, err
	}
	if !endTime.After(startTime) {
		return testUUID, ErrCompareDate
	}
	testUUID, err = s.repo.UpdateTest(ctx, userID, testUUID, test)
	return testUUID, err
}

func (s *CalendarService) Delete(ctx context.Context, userID int, testUUID string) (string, error) {
	var err error = nil
	s.msgChan <- models.TestDeleteMessage{TestUID: testUUID, UserID: userID}
	return testUUID, err
}

func (s *CalendarService) Ping(ctx context.Context) error {
	err := s.repo.Ping(ctx)
	return err
}

func (s *CalendarService) flushMessages() {
	ticker := time.NewTicker(10 * time.Second)
	var messages []models.TestDeleteMessage
	for {
		select {
		case msg := <-s.msgChan:
			messages = append(messages, msg)
		case <-ticker.C:
			ctxUUID := uuid.New().String()
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			ctxVal := context.WithValue(ctx, "UUID", ctxUUID)
			for i := range messages {
				if i == len(messages)-1 {
					ctxVal = context.WithValue(ctxVal, "isLastReq", true)
				}
				_, err := s.repo.SetDelete(ctxVal, messages[i].UserID, messages[i].TestUID)
				if err != nil {
					continue
				}
			}
			messages = nil
			cancel()
		}
	}
}

func (s *CalendarService) scheduleTests() {
	semaphore := make(chan struct{}, s.config.IntegrationRL)
	var wg sync.WaitGroup
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		tests, err := s.repo.GetTests(ctx, 0, true)
		cancel()
		if err != nil {
			logger.Log.Error("Error get tests in thread", zap.String("error", err.Error()))
		} else {
			for i := range tests {
				startTime, err := s.parseTime(tests[i].StartTime)
				if err != nil {
					logger.Log.Error("Error parse time in thread", zap.String("error", err.Error()))
				} else {
					if time.Now().After(startTime) {
						semaphore <- struct{}{}
						wg.Add(1)
						go func(testUUID string) {
							defer wg.Done()
							defer func() { <-semaphore }()
							s.startTest(testUUID)
						}(tests[i].TestUID)
					}
				}
			}
			wg.Wait()
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *CalendarService) startTest(testUUID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	test, err := s.repo.GetTest(ctx, 0, testUUID)
	reqData, err := json.Marshal(models.ReqStartTest{TestName: test.TestName, TPS: test.TPS, AdditionalParams: test.AdditionalParams})
	if err != nil {
		logger.Log.Error("Error start test in thread",
			zap.String("testUUID", testUUID),
			zap.String("error", err.Error()),
		)
		return
	}
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"http://"+s.config.IntegrationDsn,
		bytes.NewBuffer(reqData),
	)
	if err != nil {
		logger.Log.Error("Failed to create HTTP request in thread",
			zap.String("testUUID", testUUID),
			zap.String("error", err.Error()),
		)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error("Error on response in thread",
			zap.String("testUUID", testUUID),
			zap.String("error", err.Error()),
		)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Log.Error("Test start fail in thread",
			zap.String("testUUID", testUUID),
			zap.String("statusCode", resp.Status),
		)
		return
	} else {
		logger.Log.Info("Test success started in thread",
			zap.String("testUUID", testUUID),
			zap.String("statusCode", resp.Status),
		)
	}
	testUUID, err = s.repo.SetTestStarted(ctx, 0, testUUID)
	if err != nil {
		logger.Log.Error("Set test started faild in DB in thread",
			zap.String("testUUID", testUUID),
			zap.String("error", err.Error()),
		)
	}
}

func (s *CalendarService) parseTime(value string) (time.Time, error) {
	layout := "2006-01-02T15:04:05+03:00"
	t, err := time.Parse(layout, value)
	if err != nil {
		return time.Now(), ErrParsingDate
	}
	return t, nil
}

func (s *CalendarService) countEndTime(startTime time.Time, duration int) string {
	layout := "2006-01-02T15:04:05+03:00"
	endTime := startTime.Add(time.Duration(duration) * time.Minute)
	endTimeStr := endTime.Format(layout)
	return endTimeStr
}

func (s *CalendarService) generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *CalendarService) getKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(ENCRYPT_KEY_B64)
}

func (s *CalendarService) encryptPass(pass string) (string, error) {
	src := []byte(pass)
	key, err := s.getKey()
	if err != nil {
		return "", err
	}
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}
	nonce, err := s.generateRandom(aesgcm.NonceSize())
	if err != nil {
		return "", err
	}
	dst := aesgcm.Seal(nil, nonce, src, nil)
	dstWithNonce := append(nonce, dst...)
	encodedToString := base64.StdEncoding.EncodeToString(dstWithNonce)
	return encodedToString, nil
}

func (s *CalendarService) decryptPass(pass string) (string, error) {
	passBytes, err := base64.StdEncoding.DecodeString(pass)
	if err != nil {
		return "", err
	}
	key, err := s.getKey()
	if err != nil {
		return "", err
	}
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}
	nonceSize := aesgcm.NonceSize()
	decrNonce := passBytes[:nonceSize]
	ciphertext := passBytes[nonceSize:]
	decryptedPass, err := aesgcm.Open(nil, decrNonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(decryptedPass), nil
}
