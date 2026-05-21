package main

import (
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/dimalewshin98-glitch/LTCalendar/internal/config"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/handler"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/mocks"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

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

// func TestApiShortenBatchHandler(t *testing.T) {
// 	type want struct {
// 		contentType     string
// 		statusCode      int
// 		response        string
// 		contentEncoding string
// 	}
// 	tests := []struct {
// 		name            string
// 		contentType     string
// 		acceptEncoding  string
// 		contentEncoding string
// 		body            string
// 		request         string
// 		requestType     string
// 		want            want
// 	}{
// 		{
// 			name:        "test 1 | Success",
// 			contentType: "application/json",
// 			body: `[
// 						{
// 							"correlation_id": "aaa",
// 							"original_url": "urlAA"
// 						},
// 						{
// 							"correlation_id": "aaa",
// 							"original_url": "urlAA"
// 						}
// 					] `,
// 			want: want{
// 				contentType: "application/json",
// 				statusCode:  201,
// 				response:    `[{"correlation_id":"aaa","short_url":"http://localhost:8080/bjjBrD"},{"correlation_id":"aaa","short_url":"http://localhost:8080/bjjBrD"}]` + "\n",
// 			},
// 			request:     "/api/shorten/batch",
// 			requestType: "POST",
// 		},
// 		{
// 			name:            "test 2 | Success | With encodintg",
// 			contentType:     "application/json",
// 			acceptEncoding:  "gzip",
// 			contentEncoding: "gzip",
// 			body: `[
// 						{
// 							"correlation_id": "aaa",
// 							"original_url": "urlAA"
// 						},
// 						{
// 							"correlation_id": "aaa",
// 							"original_url": "urlAA"
// 						}
// 					] `,
// 			want: want{
// 				contentType:     "application/json",
// 				statusCode:      201,
// 				response:        `[{"correlation_id":"aaa","short_url":"http://localhost:8080/bjjBrD"},{"correlation_id":"aaa","short_url":"http://localhost:8080/bjjBrD"}]` + "\n",
// 				contentEncoding: "gzip",
// 			},
// 			request:     "/api/shorten/batch",
// 			requestType: "POST",
// 		},
// 		{
// 			name:        "test 3 | Unsuccess | Request type error",
// 			contentType: "application/json",
// 			body:        `{"url": "https://mockedurl.com"}`,
// 			want: want{
// 				contentType: "",
// 				statusCode:  400,
// 				response:    "Method not allowed\n",
// 			},
// 			request:     "/api/shorten/batch",
// 			requestType: "GET",
// 		},
// 		{
// 			name:        "test 4 | Unsuccess | Content-Type error",
// 			contentType: "text/html",
// 			body:        `{"url": "https://mockedurl.com"}`,
// 			want: want{
// 				contentType: "text/plain; charset=utf-8",
// 				statusCode:  400,
// 				response:    "Content-Type not allowed\n",
// 			},
// 			request:     "/api/shorten/batch",
// 			requestType: "POST",
// 		},
// 		{
// 			name:        "test 5 | Unsuccess | Batch is empty",
// 			contentType: "application/json",
// 			body:        `[]`,
// 			want: want{
// 				contentType: "text/plain; charset=utf-8",
// 				statusCode:  400,
// 				response:    "Batch is empty\n",
// 			},
// 			request:     "/api/shorten/batch",
// 			requestType: "POST",
// 		},
// 		{
// 			name:        "test 6 | Unsuccess | Json decode error",
// 			contentType: "application/json",
// 			body:        `{"url": 123}`,
// 			want: want{
// 				contentType: "text/plain; charset=utf-8",
// 				statusCode:  400,
// 				response:    "Json request decode error\n",
// 			},
// 			request:     "/api/shorten/batch",
// 			requestType: "POST",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			mockedRepository := mocks.NewMockRepositoryInterface(ctrl)
// 			mockedRepository.EXPECT().
// 				GetUsersID(gomock.Any()).
// 				Return([]int{1, 2, 3}, nil).
// 				AnyTimes()
// 			if tt.requestType == "POST" && tt.want.statusCode != 400 {
// 				mockedRepository.EXPECT().
// 					Store(gomock.Any(), 4, gomock.Any(), "urlAA").
// 					Return("bjjBrD", nil).
// 					Times(2)
// 			}
// 			mockedConfig := &config.Config{
// 				ServerHostPort:     "localhost:8080",
// 				ShortenURLHostPort: "http://localhost:8080",
// 			}
// 			shorterService := service.NewShorterService(mockedRepository, mockedConfig)
// 			requestsHandler := handler.NewRequestsHandler(shorterService)
// 			appHandler := http.HandlerFunc(requestsHandler.ApiShortenBatch)
// 			handlerWithMiddleware := handler.AuthMiddleware(appHandler, mockedRepository)
// 			request := httptest.NewRequest(tt.requestType, tt.request, strings.NewReader(tt.body))
// 			request.Header.Set("content-Type", tt.contentType)
// 			request.Header.Set("Accept-Encoding", tt.acceptEncoding)
// 			request.Header.Set("Content-Encoding", tt.contentEncoding)
// 			w := httptest.NewRecorder()
// 			handlerWithMiddleware.ServeHTTP(w, request)
// 			resBytes, _ := io.ReadAll(w.Body)
// 			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
// 			assert.Equal(t, tt.want.response, string(resBytes))
// 		})
// 	}
// }

// func TestStorenHandler(t *testing.T) {
// 	type want struct {
// 		contentType string
// 		statusCode  int
// 		response    string
// 	}
// 	tests := []struct {
// 		name        string
// 		contentType string
// 		body        string
// 		request     string
// 		requestType string
// 		want        want
// 	}{
// 		{
// 			name:        "test 1 | Success",
// 			contentType: "text/plain",
// 			body:        "urlAA",
// 			want: want{
// 				contentType: "text/plain",
// 				statusCode:  201,
// 				response:    "http://localhost:8080/bjjBrD",
// 			},
// 			request:     "/",
// 			requestType: "POST",
// 		},
// 		{
// 			name:        "test 2 | Unsuccess | Request type error",
// 			contentType: "text/plain",
// 			body:        "https://mockedurl.com",
// 			want: want{
// 				contentType: "text/plain; charset=utf-8",
// 				statusCode:  400,
// 				response:    "Method not allowed\n",
// 			},
// 			request:     "/",
// 			requestType: "GET",
// 		},
// 		{
// 			name:        "test 3 | Unsuccess | Content-Type error",
// 			contentType: "text/html",
// 			body:        "https://mockedurl.com",
// 			want: want{
// 				contentType: "text/plain; charset=utf-8",
// 				statusCode:  400,
// 				response:    "Content-Type not allowed\n",
// 			},
// 			request:     "/",
// 			requestType: "POST",
// 		},
// 		{
// 			name:        "test 4 | Unsuccess | URL is empty",
// 			contentType: "text/plain",
// 			body:        "",
// 			want: want{
// 				contentType: "text/plain; charset=utf-8",
// 				statusCode:  400,
// 				response:    "URL is empty\n",
// 			},
// 			request:     "/",
// 			requestType: "POST",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			mockedRepository := mocks.NewMockRepositoryInterface(ctrl)
// 			mockedRepository.EXPECT().
// 				GetUsersID(gomock.Any()).
// 				Return([]int{1, 2, 3}, nil).
// 				AnyTimes()
// 			if tt.requestType == "POST" && tt.want.statusCode != 400 {
// 				mockedRepository.EXPECT().
// 					Store(gomock.Any(), 4, gomock.Any(), "urlAA").
// 					Return("bjjBrD", nil).
// 					Times(1)
// 			}
// 			mockedConfig := &config.Config{
// 				ServerHostPort:     "localhost:8080",
// 				ShortenURLHostPort: "http://localhost:8080",
// 			}
// 			shorterService := service.NewShorterService(mockedRepository, mockedConfig)
// 			requestsHandler := handler.NewRequestsHandler(shorterService)
// 			appHandler := http.HandlerFunc(requestsHandler.Shorten)
// 			handlerWithMiddleware := handler.AuthMiddleware(appHandler, mockedRepository)
// 			request := httptest.NewRequest(tt.requestType, tt.request, strings.NewReader(tt.body))
// 			request.Header.Set("content-Type", tt.contentType)
// 			w := httptest.NewRecorder()
// 			handlerWithMiddleware.ServeHTTP(w, request)
// 			resBytes, _ := io.ReadAll(w.Body)
// 			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
// 			assert.Equal(t, tt.want.response, string(resBytes))
// 			assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"))
// 		})
// 	}
// }

// func TestGetURLHandler(t *testing.T) {
// 	type want struct {
// 		statusCode int
// 		response   string
// 		location   string
// 	}
// 	tests := []struct {
// 		name        string
// 		contentType string
// 		request     string
// 		requestType string
// 		want        want
// 	}{
// 		{
// 			name:        "test 1 | Success",
// 			contentType: "text/plain",
// 			want: want{
// 				statusCode: 307,
// 				response:   "",
// 				location:   "https://mockedurl.com",
// 			},
// 			request:     "/AbCdEf",
// 			requestType: "GET",
// 		},
// 		{
// 			name:        "test 2 | Unsuccess | Request type error",
// 			contentType: "text/plain",
// 			want: want{
// 				statusCode: 400,
// 				response:   "Method not allowed\n",
// 				location:   "",
// 			},
// 			request:     "/AbCdEf",
// 			requestType: "POST",
// 		},
// 		{
// 			name:        "test 3 | Unsuccess | Content-Type error",
// 			contentType: "text/html",
// 			want: want{
// 				statusCode: 400,
// 				response:   "Content-Type not allowed\n",
// 				location:   "",
// 			},
// 			request:     "/AbCdEf",
// 			requestType: "GET",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			mockedRepository := mocks.NewMockRepositoryInterface(ctrl)
// 			mockedRepository.EXPECT().
// 				GetUsersID(gomock.Any()).
// 				Return([]int{1, 2, 3}, nil).
// 				AnyTimes()
// 			if tt.requestType == "GET" && tt.want.statusCode != 400 {
// 				mockedRepository.EXPECT().
// 					Get(gomock.Any(), gomock.Any()).
// 					Return("https://mockedurl.com", nil).
// 					Times(1)
// 			}
// 			mockedConfig := &config.Config{
// 				ServerHostPort:     "localhost:8080",
// 				ShortenURLHostPort: "http://localhost:8080",
// 			}
// 			shorterService := service.NewShorterService(mockedRepository, mockedConfig)
// 			requestsHandler := handler.NewRequestsHandler(shorterService)
// 			appHandler := http.HandlerFunc(requestsHandler.GetURL)
// 			handlerWithMiddleware := handler.AuthMiddleware(appHandler, mockedRepository)
// 			request := httptest.NewRequest(tt.requestType, tt.request, strings.NewReader(""))
// 			request.Header.Set("content-Type", tt.contentType)
// 			w := httptest.NewRecorder()
// 			handlerWithMiddleware.ServeHTTP(w, request)
// 			resBytes, _ := io.ReadAll(w.Body)
// 			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
// 			assert.Equal(t, tt.want.response, string(resBytes))
// 			assert.Equal(t, tt.want.location, w.Header().Get("Location"))
// 		})
// 	}
// }

// func TestGetUserURLsHandler(t *testing.T) {
// 	type want struct {
// 		statusCode int
// 		response   string
// 		location   string
// 	}
// 	tests := []struct {
// 		name        string
// 		contentType string
// 		request     string
// 		requestType string
// 		want        want
// 	}{
// 		{
// 			name: "test 1 | Success",
// 			want: want{
// 				statusCode: 200,
// 				response:   `[{"short_url":"a","original_url":"b"},{"short_url":"c","original_url":"d"}]` + "\n",
// 			},
// 			request:     "/api/user/urls",
// 			requestType: "GET",
// 		},
// 		{
// 			name:        "test 2 | Unsuccess | No content",
// 			contentType: "text/html",
// 			want: want{
// 				statusCode: 204,
// 				response:   "",
// 				location:   "",
// 			},
// 			request:     "/api/user/urls",
// 			requestType: "GET",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			mockedRepository := mocks.NewMockRepositoryInterface(ctrl)
// 			mockedRepository.EXPECT().
// 				GetUsersID(gomock.Any()).
// 				Return([]int{1, 2, 3}, nil).
// 				AnyTimes()
// 			if tt.requestType == "GET" && tt.want.statusCode != 400 {
// 				var res models.ApiUserUrlsRes
// 				if tt.want.statusCode == 200 {
// 					res = append(res, models.UserUrlRes{ShortURL: "a", OriginalURL: "b"})
// 					res = append(res, models.UserUrlRes{ShortURL: "c", OriginalURL: "d"})
// 				}
// 				mockedRepository.EXPECT().
// 					GetUserUrls(gomock.Any(), gomock.Any()).
// 					Return(res, nil).
// 					Times(1)
// 			}
// 			mockedConfig := &config.Config{
// 				ServerHostPort:     "localhost:8080",
// 				ShortenURLHostPort: "http://localhost:8080",
// 			}
// 			shorterService := service.NewShorterService(mockedRepository, mockedConfig)
// 			requestsHandler := handler.NewRequestsHandler(shorterService)
// 			appHandler := http.HandlerFunc(requestsHandler.ApiUserUrls)
// 			handlerWithMiddleware := handler.AuthMiddleware(appHandler, mockedRepository)
// 			request := httptest.NewRequest(tt.requestType, tt.request, strings.NewReader(""))
// 			w := httptest.NewRecorder()
// 			handlerWithMiddleware.ServeHTTP(w, request)
// 			resBytes, _ := io.ReadAll(w.Body)
// 			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
// 			assert.Equal(t, tt.want.response, string(resBytes))
// 		})
// 	}
// }

// func TestDeleteURLsHandler(t *testing.T) {
// 	type want struct {
// 		contentType     string
// 		statusCode      int
// 		response        string
// 		contentEncoding string
// 	}
// 	tests := []struct {
// 		name            string
// 		contentType     string
// 		acceptEncoding  string
// 		contentEncoding string
// 		body            string
// 		request         string
// 		requestType     string
// 		want            want
// 	}{
// 		{
// 			name:        "test 1 | Success",
// 			contentType: "application/json",
// 			body:        `["aaa"]`,
// 			want: want{
// 				contentType: "application/json",
// 				statusCode:  202,
// 				response:    ``,
// 			},
// 			request:     "/api/user/urls",
// 			requestType: "DELETE",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			mockedRepository := mocks.NewMockRepositoryInterface(ctrl)
// 			mockedRepository.EXPECT().
// 				GetUsersID(gomock.Any()).
// 				Return([]int{1, 2, 3}, nil).
// 				AnyTimes()
// 			mockedRepository.EXPECT().
// 				SetDelete(gomock.Any(), 3, "aaa").
// 				Return("a", nil).
// 				AnyTimes()
// 			mockedConfig := &config.Config{
// 				ServerHostPort:     "localhost:8080",
// 				ShortenURLHostPort: "http://localhost:8080",
// 			}
// 			shorterService := service.NewShorterService(mockedRepository, mockedConfig)
// 			requestsHandler := handler.NewRequestsHandler(shorterService)
// 			appHandler := http.HandlerFunc(requestsHandler.Delete)
// 			handlerWithMiddleware := handler.AuthMiddleware(appHandler, mockedRepository)
// 			request := httptest.NewRequest(tt.requestType, tt.request, strings.NewReader(tt.body))
// 			request.Header.Set("content-Type", tt.contentType)
// 			w := httptest.NewRecorder()
// 			handlerWithMiddleware.ServeHTTP(w, request)
// 			resBytes, _ := io.ReadAll(w.Body)
// 			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
// 			assert.Equal(t, tt.want.response, string(resBytes))
// 		})
// 	}
// }
// func TestApiStorenHandler(t *testing.T) {
// 	type want struct {
// 		contentType     string
// 		statusCode      int
// 		response        string
// 		contentEncoding string
// 	}
// 	tests := []struct {
// 		name            string
// 		contentType     string
// 		acceptEncoding  string
// 		contentEncoding string
// 		body            string
// 		request         string
// 		requestType     string
// 		want            want
// 	}{
// 		{
// 			name:        "test 1 | Success",
// 			contentType: "application/json",
// 			body:        `{"url": "https://mockedurl.com"}`,
// 			want: want{
// 				contentType: "application/json",
// 				statusCode:  201,
// 				response:    `{"result":"http://localhost:8080/AbCdEf"}` + "\n",
// 			},
// 			request:     "/api/shorten",
// 			requestType: "POST",
// 		},
// 		{
// 			name:            "test 2 | Success | With encodintg",
// 			contentType:     "application/json",
// 			acceptEncoding:  "gzip",
// 			contentEncoding: "gzip",
// 			body:            `{"url": "https://mockedurl.com"}`,
// 			want: want{
// 				contentType:     "application/json",
// 				statusCode:      201,
// 				response:        `{"result":"http://localhost:8080/AbCdEf"}` + "\n",
// 				contentEncoding: "gzip",
// 			},
// 			request:     "/api/shorten",
// 			requestType: "POST",
// 		},
// 		{
// 			name:        "test 3 | Unsuccess | Request type error",
// 			contentType: "application/json",
// 			body:        `{"url": "https://mockedurl.com"}`,
// 			want: want{
// 				contentType: "",
// 				statusCode:  405,
// 				response:    "",
// 			},
// 			request:     "/api/shorten",
// 			requestType: "GET",
// 		},
// 		{
// 			name:        "test 4 | Unsuccess | Content-Type error",
// 			contentType: "text/html",
// 			body:        `{"url": "https://mockedurl.com"}`,
// 			want: want{
// 				contentType: "text/plain; charset=utf-8",
// 				statusCode:  400,
// 				response:    "Content-Type not allowed\n",
// 			},
// 			request:     "/api/shorten",
// 			requestType: "POST",
// 		},
// 		{
// 			name:        "test 5 | Unsuccess | URL is empty",
// 			contentType: "application/json",
// 			body:        `{"url": ""}`,
// 			want: want{
// 				contentType: "text/plain; charset=utf-8",
// 				statusCode:  400,
// 				response:    "URL is empty\n",
// 			},
// 			request:     "/api/shorten",
// 			requestType: "POST",
// 		},
// 		{
// 			name:        "test 6 | Unsuccess | Json decode error",
// 			contentType: "application/json",
// 			body:        `{"url": 123}`,
// 			want: want{
// 				contentType: "text/plain; charset=utf-8",
// 				statusCode:  400,
// 				response:    "Json request decode error\n",
// 			},
// 			request:     "/api/shorten",
// 			requestType: "POST",
// 		},
// 	}
// 	repository := repository.NewInmemoryRepository()
// 	mockedRepository := &MockedInmemoryRepository{repository}
// 	mockedConfig := &config.Config{
// 		ServerHostPort:     "localhost:8080",
// 		ShortenURLHostPort: "http://localhost:8080",
// 	}
// 	app := NewApp(mockedRepository, *mockedConfig)
// 	appHandler := app.GetHandler()
// 	srv := httptest.NewServer((handler.GzipMiddleware(handler.AuthMiddleware(appHandler, repository))))
// 	defer srv.Close()
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var buf *bytes.Buffer
// 			if tt.contentEncoding == "gzip" {
// 				var b bytes.Buffer
// 				zb := gzip.NewWriter(&b)
// 				zb.Write([]byte(tt.body))
// 				zb.Close()
// 				buf = &b
// 			} else {
// 				buf = bytes.NewBufferString(tt.body)
// 			}
// 			r := httptest.NewRequest(tt.requestType, srv.URL+tt.request, buf)
// 			r.RequestURI = ""
// 			r.Header.Set("Content-Type", tt.contentType)
// 			r.Header.Set("Accept-Encoding", tt.acceptEncoding)
// 			r.Header.Set("Content-Encoding", tt.contentEncoding)
// 			resp, _ := http.DefaultClient.Do(r)
// 			var resBytes []byte
// 			if tt.acceptEncoding == "gzip" {
// 				zr, _ := gzip.NewReader(resp.Body)
// 				resBytes, _ = io.ReadAll(zr)
// 			} else {
// 				resBytes, _ = io.ReadAll(resp.Body)
// 			}
// 			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
// 			assert.Equal(t, tt.want.response, string(resBytes))
// 			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
// 			assert.Equal(t, tt.want.contentEncoding, resp.Header.Get("Content-Encoding"))
// 			defer resp.Body.Close()
// 		})
// 	}
// }
