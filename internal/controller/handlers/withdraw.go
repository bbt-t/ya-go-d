package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bbt-t/ya-go-d/internal/adapter/storage"
	"github.com/bbt-t/ya-go-d/internal/entity"
	luhn "github.com/bbt-t/ya-go-d/pkg/luhnalgorithm"
)

func (g GophermartHandler) wd(w http.ResponseWriter, r *http.Request) {
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
			fmt.Sprintf("wrong body: %v", err),
			http.StatusBadRequest,
		)
		return
	}

	value := r.Context().Value("user_id")

	switch value.(type) {
	case entity.User:
		break
	default:
		log.Println("Wrong value type in context")
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	userObj := value.(entity.User)
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
		log.Println("Can't withdraw money:", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
