package handlers

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bbt-t/ya-go-d/internal/adapter/storage"
	"github.com/bbt-t/ya-go-d/internal/entity"
	luhn "github.com/bbt-t/ya-go-d/pkg/luhnalgorithm"
)

func (g GophermartHandler) order(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	if !strings.Contains(contentType, "text/plain") {
		w.Header().Set("Accept", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "wrong body: "+err.Error(), http.StatusBadRequest)
		return
	}

	orderNumber, err := strconv.Atoi(string(payload))
	if err != nil {
		http.Error(w, "wrong order number", http.StatusBadRequest)
		return
	}

	userObj, ok := r.Context().Value(entity.CtxUserKey("user_id")).(entity.User)
	if !ok {
		log.Println("Wrong value type in context")
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	order := entity.Order{
		UserID:    userObj.ID,
		Number:    orderNumber,
		Status:    "NEW",
		EventTime: time.Now(),
	}
	if !luhn.Validate(order.Number) {
		http.Error(w, "wrong order number", http.StatusUnprocessableEntity)
		return
	}

	err = g.s.AddOrder(r.Context(), order)

	if errors.Is(err, storage.ErrNumAlreadyLoaded) {
		w.WriteHeader(http.StatusOK)
		return
	}
	if errors.Is(err, storage.ErrWrongUser) {
		http.Error(w, "already loaded by another user", http.StatusConflict)
		return
	}
	if err != nil {
		log.Println("Can't add new order:", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
