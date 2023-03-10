package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/bbt-t/ya-go-d/internal/entity"
)

func (g GopherMartHandler) getBalance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	userObj, ok := r.Context().Value(entity.CtxUserKey("user_id")).(entity.User)
	if !ok {
		log.Println("Wrong value type in context")
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	user, err := g.s.GetUser(ctx, entity.SearchByID, strconv.Itoa(userObj.ID))
	if err != nil {
		log.Printf("Failed fetch balance: %+v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	balance := entity.Balance{
		Current:   user.Balance,
		Withdrawn: user.Withdrawn,
	}

	b, err := json.Marshal(balance)
	if err != nil {
		log.Printf("Failed marshalling balance: %+v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
