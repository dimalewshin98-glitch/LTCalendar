package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dimalewshin98-glitch/LTCalendar/internal/repository"
	"github.com/golang-jwt/jwt/v4"
)

var ErrNamedCookieNotPresent = errors.New("http: named cookie not present")

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const TOKEN_EXP = time.Hour * 3
const SECRET_KEY = "supersecretkey"

func BuildJWTString(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		UserID: userID,
	})
	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func CreateUserID(repo repository.RepositoryInterface) (int, error) {
	dbctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	usersID, err := repo.GetUsersID(dbctx)
	if err != nil {
		return 0, err
	}
	if len(usersID) == 0 {
		return 1, nil
	}
	maxUserID := usersID[0]
	for _, userId := range usersID {
		if userId > maxUserID {
			maxUserID = userId
		}
	}
	maxUserID += 1
	return maxUserID, nil
}

func GetUserID(tokenString string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SECRET_KEY), nil
		})
	if err != nil {
		return -1
	}
	if !token.Valid {
		return -1
	}
	return claims.UserID
}

func AuthMiddleware(h http.Handler, repo repository.RepositoryInterface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		cookie, err := r.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				userID, err := CreateUserID(repo)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				tokenString, err = BuildJWTString(userID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			tokenString = cookie.Value
		}
		userID := GetUserID(tokenString)
		switch userID {
		case 0:
			http.Error(w, "", http.StatusUnauthorized)
			return
		case -1:
			userID, err = CreateUserID(repo)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			tokenString, err = BuildJWTString(userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		http.SetCookie(w, &http.Cookie{
			Name:  "token",
			Value: tokenString,
		})
		ctx := context.WithValue(r.Context(), "userID", userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
