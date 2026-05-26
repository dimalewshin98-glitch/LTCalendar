package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dimalewshin98-glitch/LTCalendar/internal/config"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/handler"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/mocks"
	models "github.com/dimalewshin98-glitch/LTCalendar/internal/model"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/repository"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/service"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const TokenExp = time.Hour * 3
const TestSecretKey = "supersecretkey"
const TestEncryptKey = "CCHki0M3gumr8P8fs6I7IkyHHdUeNpEbV1/lpbQOtGg="

func BuildJWTString(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})
	tokenString, err := token.SignedString([]byte(TestSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func TestPingHandler(t *testing.T) {
	type want struct {
		statusCode int
		response   string
		dbResponse error
	}
	tests := []struct {
		name        string
		request     string
		requestType string
		want        want
	}{
		{
			name: "test 1 | Success",
			want: want{
				statusCode: 200,
				response:   "",
				dbResponse: nil,
			},
			request:     "/ping",
			requestType: "GET",
		},
		{
			name: "test 2 | Unsuccess | Request type error",
			want: want{
				statusCode: 400,
				response:   "Method not allowed\n",
				dbResponse: nil,
			},
			request:     "/ping",
			requestType: "POST",
		},
		{
			name: "test 3 | Unsuccess | Repo ping error",
			want: want{
				statusCode: 500,
				response:   "db connection failed\n",
				dbResponse: errors.New("db connection failed"),
			},
			request:     "/ping",
			requestType: "GET",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockedRepository := mocks.NewMockRepositoryInterface(ctrl)
			if tt.requestType == "GET" {
				mockedRepository.EXPECT().
					Ping(gomock.Any()).
					Return(tt.want.dbResponse)
			}
			mockedConfig := &config.Config{
				ServerHostPort: "localhost:8080",
			}
			shorterService := service.NewCalendarService(mockedRepository, mockedConfig)
			requestsHandler := handler.NewRequestsHandler(shorterService)
			request := httptest.NewRequest(tt.requestType, tt.request, nil)
			w := httptest.NewRecorder()
			requestsHandler.Ping(w, request)
			resBytes, _ := io.ReadAll(w.Body)
			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
			assert.Equal(t, tt.want.response, string(resBytes))
		})
	}
}

func TestRegisterUser(t *testing.T) {
	type want struct {
		contentType   string
		statusCode    int
		response      string
		dbResponse    int
		dbErrResponse error
	}
	tests := []struct {
		name            string
		contentType     string
		acceptEncoding  string
		contentEncoding string
		body            string
		request         string
		requestType     string
		want            want
	}{
		{
			name:        "test 1 | Success",
			contentType: "application/json",
			request:     "/api/register",
			requestType: "POST",
			body:        `{"login" : "aaa", "password" : "bbb"}`,
			want: want{
				contentType:   "application/json",
				statusCode:    200,
				response:      `{"user_id":666}` + "\n",
				dbResponse:    666,
				dbErrResponse: nil,
			},
		},
		{
			name:        "test 2 | Unsuccess | User exists",
			contentType: "application/json",
			request:     "/api/register",
			requestType: "POST",
			body:        `{"login" : "aaa", "password" : "bbb"}`,
			want: want{
				contentType:   "application/json",
				statusCode:    401,
				response:      `user already exists` + "\n",
				dbResponse:    0,
				dbErrResponse: repository.ErrUserAlreadyExists,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockedRepository := mocks.NewMockRepositoryInterface(ctrl)
			mockedRepository.EXPECT().
				Register(gomock.Any(), gomock.Any()).
				Return(tt.want.dbResponse, tt.want.dbErrResponse)
			mockedConfig := &config.Config{
				ServerHostPort: "localhost:8080",
				EncryptKey:     TestEncryptKey,
			}
			shorterService := service.NewCalendarService(mockedRepository, mockedConfig)
			requestsHandler := handler.NewRequestsHandler(shorterService)
			request := httptest.NewRequest(tt.requestType, tt.request, strings.NewReader(tt.body))
			request.Header.Set("content-Type", tt.contentType)
			w := httptest.NewRecorder()
			requestsHandler.RegisterUser(w, request)
			resBytes, _ := io.ReadAll(w.Body)
			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
			assert.Equal(t, tt.want.response, string(resBytes))
		})
	}
}

func TestLoginUser(t *testing.T) {
	type want struct {
		contentType        string
		statusCode         int
		response           string
		dbResponseUserID   int
		dbResponseUserPass string
		dbErrResponse      error
	}
	tests := []struct {
		name            string
		contentType     string
		acceptEncoding  string
		contentEncoding string
		body            string
		request         string
		requestType     string
		want            want
	}{
		{
			name:        "test 1 | Success",
			contentType: "application/json",
			request:     "/api/login",
			requestType: "POST",
			body:        `{"login" : "denis", "password" : "superpassword"}`,
			want: want{
				contentType:        "application/json",
				statusCode:         200,
				response:           `{"user_id":666}` + "\n",
				dbResponseUserID:   666,
				dbResponseUserPass: "mZpf6RDw3wgNfxEyjpCozf5DctwSokhs+5ynxDglA3R93+mYK3eeuvc=",
				dbErrResponse:      nil,
			},
		},
		{
			name:        "test 2 | Unsuccess | Wrong logpass",
			contentType: "application/json",
			request:     "/api/login",
			requestType: "POST",
			body:        `{"login" : "aaa", "password" : "bbb"}`,
			want: want{
				contentType:        "application/json",
				statusCode:         401,
				response:           `login or password incorrect` + "\n",
				dbResponseUserID:   666,
				dbResponseUserPass: "password",
				dbErrResponse:      repository.ErrUserLogPassIncorrect,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockedRepository := mocks.NewMockRepositoryInterface(ctrl)
			mockedRepository.EXPECT().
				Login(gomock.Any(), gomock.Any()).
				Return(tt.want.dbResponseUserID, tt.want.dbResponseUserPass, tt.want.dbErrResponse)
			mockedConfig := &config.Config{
				ServerHostPort: "localhost:8080",
				EncryptKey:     TestEncryptKey,
			}
			shorterService := service.NewCalendarService(mockedRepository, mockedConfig)
			requestsHandler := handler.NewRequestsHandler(shorterService)
			request := httptest.NewRequest(tt.requestType, tt.request, strings.NewReader(tt.body))
			request.Header.Set("content-Type", tt.contentType)
			w := httptest.NewRecorder()
			requestsHandler.LoginUser(w, request)
			resBytes, _ := io.ReadAll(w.Body)
			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
			assert.Equal(t, tt.want.response, string(resBytes))
		})
	}
}

func TestAddTest(t *testing.T) {
	type want struct {
		contentType   string
		statusCode    int
		response      string
		dbResponse    string
		dbErrResponse error
	}
	tests := []struct {
		name            string
		contentType     string
		acceptEncoding  string
		contentEncoding string
		body            string
		request         string
		requestType     string
		want            want
	}{
		{
			name:        "test 1 | Success",
			contentType: "application/json",
			request:     "/api/tests",
			requestType: "POST",
			body: `{
						"test_name": "STABILITY",
						"start_time": "2026-01-02T15:04:05+03:00",
						"duration": 12,
						"tps": 66456.666,
						"additional_params": "some some some"
					}`,
			want: want{
				contentType:   "application/json",
				statusCode:    201,
				response:      `{"test_uuid":"testuuid"}` + "\n",
				dbResponse:    "testuuid",
				dbErrResponse: nil,
			},
		},
		{
			name:        "test 2 | Unsuccess | DB error",
			contentType: "application/json",
			request:     "/api/tests",
			requestType: "POST",
			body: `{
						"test_name": "STABILITY",
						"start_time": "2026-01-02T15:04:05+03:00",
						"duration": 12,
						"tps": 66456.666,
						"additional_params": "some some some"
					}`,
			want: want{
				contentType:   "application/json",
				statusCode:    400,
				response:      `Some DB error` + "\n",
				dbResponse:    "",
				dbErrResponse: errors.New("Some DB error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockedRepository := mocks.NewMockRepositoryInterface(ctrl)
			mockedRepository.EXPECT().
				AddTest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.want.dbResponse, tt.want.dbErrResponse)
			mockedConfig := &config.Config{
				ServerHostPort: "localhost:8080",
			}
			shorterService := service.NewCalendarService(mockedRepository, mockedConfig)
			requestsHandler := handler.NewRequestsHandler(shorterService)
			appHandler := http.HandlerFunc(requestsHandler.AddTest)
			handlerWithMiddleware := handler.AuthMiddleware(appHandler, mockedRepository, TestSecretKey)
			request := httptest.NewRequest(tt.requestType, tt.request, strings.NewReader(tt.body))
			request.Header.Set("content-Type", tt.contentType)
			token, _ := BuildJWTString(666)
			cookie := &http.Cookie{
				Name:  "token",
				Value: token,
				Path:  "/api/",
			}
			request.AddCookie(cookie)
			w := httptest.NewRecorder()
			handlerWithMiddleware.ServeHTTP(w, request)
			resBytes, _ := io.ReadAll(w.Body)
			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
			assert.Equal(t, tt.want.response, string(resBytes))
		})
	}
}

func TestUpdateTest(t *testing.T) {
	type want struct {
		contentType   string
		statusCode    int
		response      string
		dbResponse    string
		dbErrResponse error
	}
	tests := []struct {
		name            string
		contentType     string
		acceptEncoding  string
		contentEncoding string
		body            string
		request         string
		requestType     string
		want            want
	}{
		{
			name:        "test 1 | Success",
			contentType: "application/json",
			request:     "/api/tests",
			requestType: "PUT",
			body: `{
						"test_name": "STABILITY",
						"start_time": "2026-01-02T15:04:05+03:00",
						"end_time": "2026-01-03T15:04:05+03:00",
						"tps": 6.6,
						"additional_params": "some some some"
					}`,
			want: want{
				contentType:   "application/json",
				statusCode:    202,
				response:      ``,
				dbResponse:    "testuuid",
				dbErrResponse: nil,
			},
		},
		{
			name:        "test 2 | Unsuccess | Test already deleted",
			contentType: "application/json",
			request:     "/api/tests",
			requestType: "PUT",
			body: `{
						"test_name": "STABILITY",
						"start_time": "2026-01-02T15:04:05+03:00",
						"end_time": "2026-01-03T15:04:05+03:00",
						"tps": 6.6,
						"additional_params": "some some some"
					}`,
			want: want{
				contentType:   "application/json",
				statusCode:    410,
				response:      ``,
				dbResponse:    "",
				dbErrResponse: repository.ErrTestDeleted,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockedRepository := mocks.NewMockRepositoryInterface(ctrl)
			mockedRepository.EXPECT().
				UpdateTest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.want.dbResponse, tt.want.dbErrResponse)
			mockedConfig := &config.Config{
				ServerHostPort: "localhost:8080",
			}
			shorterService := service.NewCalendarService(mockedRepository, mockedConfig)
			requestsHandler := handler.NewRequestsHandler(shorterService)
			appHandler := http.HandlerFunc(requestsHandler.UpdateTest)
			handlerWithMiddleware := handler.AuthMiddleware(appHandler, mockedRepository, TestSecretKey)
			request := httptest.NewRequest(tt.requestType, tt.request, strings.NewReader(tt.body))
			request.Header.Set("content-Type", tt.contentType)
			token, _ := BuildJWTString(666)
			cookie := &http.Cookie{
				Name:  "token",
				Value: token,
				Path:  "/api/",
			}
			request.AddCookie(cookie)
			w := httptest.NewRecorder()
			handlerWithMiddleware.ServeHTTP(w, request)
			resBytes, _ := io.ReadAll(w.Body)
			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
			assert.Equal(t, tt.want.response, string(resBytes))
		})
	}
}

func TestDeleteTest(t *testing.T) {
	type want struct {
		contentType   string
		statusCode    int
		response      string
		dbResponse    string
		dbErrResponse error
	}
	tests := []struct {
		name            string
		contentType     string
		acceptEncoding  string
		contentEncoding string
		body            string
		request         string
		requestType     string
		want            want
	}{
		{
			name:        "test 1 | Success",
			contentType: "application/json",
			request:     "/api/tests/sometestuuid",
			requestType: "DELETE",
			body:        ``,
			want: want{
				contentType:   "application/json",
				statusCode:    202,
				response:      ``,
				dbResponse:    "testuuid",
				dbErrResponse: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockedRepository := mocks.NewMockRepositoryInterface(ctrl)
			mockedConfig := &config.Config{
				ServerHostPort: "localhost:8080",
			}
			shorterService := service.NewCalendarService(mockedRepository, mockedConfig)
			requestsHandler := handler.NewRequestsHandler(shorterService)
			appHandler := http.HandlerFunc(requestsHandler.DeleteTest)
			handlerWithMiddleware := handler.AuthMiddleware(appHandler, mockedRepository, TestSecretKey)
			request := httptest.NewRequest(tt.requestType, tt.request, strings.NewReader(tt.body))
			request.Header.Set("content-Type", tt.contentType)
			token, _ := BuildJWTString(666)
			cookie := &http.Cookie{
				Name:  "token",
				Value: token,
				Path:  "/api/",
			}
			request.AddCookie(cookie)
			w := httptest.NewRecorder()
			handlerWithMiddleware.ServeHTTP(w, request)
			resBytes, _ := io.ReadAll(w.Body)
			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
			assert.Equal(t, tt.want.response, string(resBytes))
		})
	}
}

func TestGetTest(t *testing.T) {
	type want struct {
		contentType   string
		statusCode    int
		response      string
		dbResponse    models.ApiGetTestRes
		dbErrResponse error
	}
	tests := []struct {
		name            string
		contentType     string
		acceptEncoding  string
		contentEncoding string
		body            string
		request         string
		requestType     string
		want            want
	}{
		{
			name:        "test 1 | Success",
			contentType: "application/json",
			request:     "/api/tests/sometestuuid",
			requestType: "GET",
			body:        ``,
			want: want{
				contentType: "application/json",
				statusCode:  200,
				response:    `{"test_name":"STABILITY","start_time":"2026-01-02T15:04:05+03:00","end_time":"2026-01-02T03:20:05+03:00","tps":666.66,"additional_params":"some some some","is_started":false}` + "\n",
				dbResponse: models.ApiGetTestRes{
					TestName:         "STABILITY",
					StartTime:        "2026-01-02T15:04:05+03:00",
					EndTime:          "2026-01-02T03:20:05+03:00",
					TPS:              666.66,
					AdditionalParams: "some some some",
					IsStarted:        false},
				dbErrResponse: nil,
			},
		},
		{
			name:        "test 2 | Unsuccess | Test not exists",
			contentType: "application/json",
			request:     "/api/tests/sometestuuid",
			requestType: "GET",
			body:        ``,
			want: want{
				contentType:   "application/json",
				statusCode:    404,
				response:      ``,
				dbResponse:    models.ApiGetTestRes{},
				dbErrResponse: repository.ErrTestNotExists,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockedRepository := mocks.NewMockRepositoryInterface(ctrl)
			mockedRepository.EXPECT().
				GetTest(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.want.dbResponse, tt.want.dbErrResponse)
			mockedConfig := &config.Config{
				ServerHostPort: "localhost:8080",
			}
			shorterService := service.NewCalendarService(mockedRepository, mockedConfig)
			requestsHandler := handler.NewRequestsHandler(shorterService)
			appHandler := http.HandlerFunc(requestsHandler.GetTest)
			handlerWithMiddleware := handler.AuthMiddleware(appHandler, mockedRepository, TestSecretKey)
			request := httptest.NewRequest(tt.requestType, tt.request, nil)
			request.Header.Set("content-Type", tt.contentType)
			token, _ := BuildJWTString(666)
			cookie := &http.Cookie{
				Name:  "token",
				Value: token,
				Path:  "/api/",
			}
			request.AddCookie(cookie)
			w := httptest.NewRecorder()
			handlerWithMiddleware.ServeHTTP(w, request)
			resBytes, _ := io.ReadAll(w.Body)
			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
			assert.Equal(t, tt.want.response, string(resBytes))
		})
	}
}

func TestGetTests(t *testing.T) {
	type want struct {
		contentType   string
		statusCode    int
		response      string
		dbResponse    models.ApiGetTestsRes
		dbErrResponse error
	}
	tests := []struct {
		name            string
		contentType     string
		acceptEncoding  string
		contentEncoding string
		body            string
		request         string
		requestType     string
		want            want
	}{
		{
			name:        "test 1 | Success",
			contentType: "application/json",
			request:     "/api/tests/sometestuuid",
			requestType: "GET",
			body:        ``,
			want: want{
				contentType: "application/json",
				statusCode:  200,
				response:    `[{"test_uuid":"","test_name":"STABILITY","start_time":"2026-01-02T15:04:05+03:00","is_started":false}]` + "\n",
				dbResponse: models.ApiGetTestsRes{{
					TestName:  "STABILITY",
					StartTime: "2026-01-02T15:04:05+03:00",
					IsStarted: false}},
				dbErrResponse: nil,
			},
		},
		{
			name:        "test 2 | Unsuccess | DB error",
			contentType: "application/json",
			request:     "/api/tests/sometestuuid",
			requestType: "GET",
			body:        ``,
			want: want{
				contentType:   "application/json",
				statusCode:    400,
				response:      `Some DB error` + "\n",
				dbResponse:    models.ApiGetTestsRes{},
				dbErrResponse: errors.New("Some DB error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockedRepository := mocks.NewMockRepositoryInterface(ctrl)
			mockedRepository.EXPECT().
				GetTests(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.want.dbResponse, tt.want.dbErrResponse)
			mockedConfig := &config.Config{
				ServerHostPort: "localhost:8080",
			}
			shorterService := service.NewCalendarService(mockedRepository, mockedConfig)
			requestsHandler := handler.NewRequestsHandler(shorterService)
			appHandler := http.HandlerFunc(requestsHandler.GetTests)
			handlerWithMiddleware := handler.AuthMiddleware(appHandler, mockedRepository, TestSecretKey)
			request := httptest.NewRequest(tt.requestType, tt.request, nil)
			request.Header.Set("content-Type", tt.contentType)
			token, _ := BuildJWTString(666)
			cookie := &http.Cookie{
				Name:  "token",
				Value: token,
				Path:  "/api/",
			}
			request.AddCookie(cookie)
			w := httptest.NewRecorder()
			handlerWithMiddleware.ServeHTTP(w, request)
			resBytes, _ := io.ReadAll(w.Body)
			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
			assert.Equal(t, tt.want.response, string(resBytes))
		})
	}
}
