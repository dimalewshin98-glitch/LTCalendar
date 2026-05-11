package logger

import (
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return err
}

type loggedResponseWriter struct {
	http.ResponseWriter
	statusCode int
	bodySize   int
}

func (rw *loggedResponseWriter) WriteHeader(code int) {
	rw.ResponseWriter.WriteHeader(code)
	rw.statusCode = code
}

func (lw *loggedResponseWriter) Write(data []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(data)
	lw.bodySize += size
	return size, err
}

func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		lw := &loggedResponseWriter{ResponseWriter: w, statusCode: 200, bodySize: 0}
		h.ServeHTTP(lw, r)
		duration := strconv.FormatInt(time.Since(start).Milliseconds(), 10)
		Log.Info("got incoming HTTP request",
			zap.String("uri", uri),
			zap.String("method", method),
			zap.String("duration", duration+"ms"),
		)
		Log.Info("send HTTP response",
			zap.String("status", strconv.Itoa(lw.statusCode)),
			zap.String("body size", strconv.Itoa(lw.bodySize)),
		)
	})

}
