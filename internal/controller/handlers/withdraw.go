package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bbt-t/ya-go-d/internal/adapter/storage"
	"github.com/bbt-t/ya-go-d/internal/entity"
	luhn "github.com/bbt-t/ya-go-d/pkg/luhnalgorithm"
)

func (g GopherMartHandler) wd(w http.ResponseWriter, r *http.Request) {
	var withdrawal entity.Withdraw
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

	if err = json.Unmarshal(payload, &withdrawal); err != nil {
		http.Error(
			w,
			strings.Join([]string{"wrong payload:", err.Error()}, " "),
			http.StatusBadRequest,
		)
		return
	}

	userObj, ok := r.Context().Value(entity.CtxUserKey("user_id")).(entity.User)
	if !ok {
		log.Println("Wrong value type in context")
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if !luhn.Validate(withdrawal.Order) {
		http.Error(w, "wrong order number", http.StatusUnprocessableEntity)
		return
	}

	err = g.s.Withdraw(r.Context(), userObj, withdrawal)
	if errors.Is(err, storage.ErrNoEnoughBalance) {
		http.Error(w, "Insufficient balance", http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		log.Printf("Can't withdraw money: %+v\n", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
