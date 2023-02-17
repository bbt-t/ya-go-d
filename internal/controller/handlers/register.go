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
)

func (g GophermartHandler) reg(w http.ResponseWriter, r *http.Request) {
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

	http.SetCookie(
		w,
		&http.Cookie{
			Name:    "userCookie",
			Value:   hex.EncodeToString(pkg.MakeCookie(userObj.ID, g.cfg.SecretKey)),
			Expires: time.Now().Add(365 * 24 * time.Hour),
			Path:    "/",
		},
	)
	w.WriteHeader(http.StatusOK)
}
