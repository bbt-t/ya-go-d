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
)

func (g GophermartHandler) reg(w http.ResponseWriter, r *http.Request) {
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

	userObj := entity.User{}

	if err = json.Unmarshal(payload, &userObj); err != nil {
		http.Error(w, "wrong body: "+err.Error(), http.StatusBadRequest)
		return
	}

	userObj.ID, err = g.s.NewUser(r.Context(), userObj)

	if errors.Is(err, storage.ErrExists) {
		http.Error(w, "login already exists", http.StatusConflict)
		return
	}
	if err != nil {
		log.Println("Failed add user:", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	createdCookie := pkg.MakeCookie(userObj.ID, g.cfg.SecretKey)
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{
		Name:    "userCookie",
		Value:   hex.EncodeToString(createdCookie),
		Expires: expiration,
		Path:    "/",
	}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)
}
