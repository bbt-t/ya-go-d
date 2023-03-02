package handlers

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bbt-t/ya-go-d/internal/adapter/storage"
	"github.com/bbt-t/ya-go-d/internal/entity"
	"github.com/bbt-t/ya-go-d/pkg"

	"golang.org/x/crypto/bcrypt"
)

func (g GopherMartHandler) login(w http.ResponseWriter, r *http.Request) {
	var userObj entity.User

	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		w.Header().Set("Accept", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, errBody := io.ReadAll(r.Body)
	defer r.Body.Close()

	if errBody != nil {
		http.Error(
			w,
			strings.Join([]string{"wrong body:", errBody.Error()}, " "),
			http.StatusBadRequest,
		)
		return
	}
	if err := json.Unmarshal(payload, &userObj); err != nil {
		http.Error(
			w,
			strings.Join([]string{"wrong payload:", err.Error()}, " "),
			http.StatusBadRequest,
		)
		return
	}

	user, errGet := g.s.GetUser(r.Context(), entity.SearchByLogin, userObj.Login)
	if errors.Is(errGet, storage.ErrNotFound) {
		http.Error(w, "login doesn't exists", http.StatusUnauthorized)
		return
	}
	if errGet != nil {
		log.Printf("Failed get user: %+v\n", errGet)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(strings.Join([]string{userObj.Login, userObj.Password}, "")),
	); err != nil {
		http.Error(w, "wrong credentials", http.StatusUnauthorized)
		return
	}

	http.SetCookie(
		w,
		&http.Cookie{
			Name:    "userCookie",
			Value:   hex.EncodeToString(pkg.MakeCookie(user.ID, g.cfg.SecretKey)),
			Expires: time.Now().Add(365 * 24 * time.Hour),
			Path:    "/",
		},
	)
	w.WriteHeader(http.StatusOK)
}
