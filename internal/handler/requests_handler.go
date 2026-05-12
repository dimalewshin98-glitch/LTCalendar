package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	models "github.com/dimalewshin98-glitch/LTCalendar/internal/model"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/service"
)

type RequestsHandler struct {
	service service.ServiceInterface
}

func NewRequestsHandler(service service.ServiceInterface) *RequestsHandler {
	return &RequestsHandler{
		service: service,
	}
}

func (s *RequestsHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	err := s.service.Ping(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *RequestsHandler) AddTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type not allowed", http.StatusBadRequest)
		return
	}
	var req models.ApiAddTestReq
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := dec.Decode(&req)
	if err != nil {
		http.Error(w, "Json request decode error", http.StatusBadRequest)
		return
	}
	testUUID, err := s.service.AddTest(ctx, userID, req)
	resHeader := http.StatusCreated
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res := models.ApiAddTestRes{
		TestUID: testUUID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resHeader)
	enc := json.NewEncoder(w)
	err = enc.Encode(res)
	if err != nil {
		http.Error(w, "Json response encode error", http.StatusBadRequest)
		return
	}
}

func (s *RequestsHandler) GetTests(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	tests, err := s.service.GetTests(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(tests)
	if err != nil {
		http.Error(w, "Json response encode error", http.StatusBadRequest)
		return
	}
}
