package controller

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/bbt-t/ya-go-d/internal/config"
	"github.com/bbt-t/ya-go-d/internal/entity"
)

func RequireAuthentication(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCookie, err := r.Cookie("userCookie")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userObj := entity.User{}

			cookie, err := hex.DecodeString(userCookie.Value)

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			sign := append([]byte{cookie[8]}, cookie[9:40]...)
			data := append(cookie[:8], cookie[40:]...)

			h := hmac.New(sha256.New, []byte(cfg.SecretKey))
			h.Write(data)

			s := h.Sum(nil)
			if !hmac.Equal(sign, s) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userObj.ID, err = strconv.Atoi(string(cookie[40:]))
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), entity.CtxUserKey{}, userObj)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
