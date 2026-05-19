package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	models "github.com/dimalewshin98-glitch/LTCalendar/internal/model"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/repository"
	"github.com/dimalewshin98-glitch/LTCalendar/internal/service"
)

var ErrPotReqValidate = errors.New("Request validate error")

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

func (s *RequestsHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
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
	var req models.ApiLoginReq
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := dec.Decode(&req)
	if err != nil {
		http.Error(w, "Json request decode error", http.StatusBadRequest)
		return
	}
	res, err := s.service.Register(ctx, req)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("User already exists"))
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(res)
	if err != nil {
		http.Error(w, "Json response encode error", http.StatusBadRequest)
		return
	}
}

func (s *RequestsHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
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
	var req models.ApiLoginReq
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := dec.Decode(&req)
	if err != nil {
		http.Error(w, "Json request decode error", http.StatusBadRequest)
		return
	}
	res, err := s.service.Login(ctx, req)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotExists) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("User not exists"))
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	tokenString, err := BuildJWTString(res.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: tokenString,
		Path:  "/api/",
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(res)
	if err != nil {
		http.Error(w, "Json response encode error", http.StatusBadRequest)
		return
	}
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
	if len(tests) == 0 {
		http.Error(w, "", http.StatusNoContent)
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

func (s *RequestsHandler) GetTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	testUUID := r.PathValue("uuid")
	test, err := s.service.GetTest(ctx, userID, testUUID)
	if err != nil {
		if errors.Is(err, repository.ErrTestNotExists) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if errors.Is(err, repository.ErrTestDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(test)
	if err != nil {
		http.Error(w, "Json response encode error", http.StatusBadRequest)
		return
	}
}

func (s *RequestsHandler) UpdateTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	testUUID := r.PathValue("uuid")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type not allowed", http.StatusBadRequest)
		return
	}
	var req models.ApiUpdateTestReq
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := dec.Decode(&req)
	if err != nil {
		http.Error(w, "Json request decode error", http.StatusBadRequest)
		return
	}
	validate := validator.New()
	err = validate.Struct(req)
	if err != nil {
		http.Error(w, ErrPotReqValidate.Error(), http.StatusBadRequest)
		return
	}
	_, err = s.service.UpdateTest(ctx, userID, testUUID, req)
	resHeader := http.StatusAccepted
	if err != nil {
		if errors.Is(err, repository.ErrTestDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.WriteHeader(resHeader)
}

func (s *RequestsHandler) DeleteTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	testUUID := r.PathValue("uuid")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	_, err := s.service.Delete(ctx, userID, testUUID)
	resHeader := http.StatusAccepted
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resHeader)
}
