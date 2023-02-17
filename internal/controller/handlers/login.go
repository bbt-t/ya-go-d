package handlers

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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

func (g GophermartHandler) login(w http.ResponseWriter, r *http.Request) {
	var userObj entity.User
	contentType := r.Header.Get("Content-Type")

	if !strings.Contains(contentType, "application/json") {
		w.Header().Set("Accept", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "wrong body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(payload, &userObj); err != nil {
		http.Error(
			w,
			fmt.Sprintf("wrong body: %v", err),
			http.StatusBadRequest,
		)
		return
	}

	user, err := g.s.GetUser(r.Context(), entity.SearchByLogin, userObj.Login)
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(w, "login doesn't exists", http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.Println("Failed get user:", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(fmt.Sprintf("%v%v", userObj.Login, userObj.Password)),
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
