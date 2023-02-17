package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bbt-t/ya-go-d/internal/entity"
)

func (g GophermartHandler) getBalance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	value := r.Context().Value(entity.CtxUserKey{})

	switch value.(type) {
	case entity.User:
		break
	default:
		log.Println("Wrong value type in context")
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	userObj := value.(entity.User)
	user, err := g.s.GetUser(ctx, entity.SearchByID, fmt.Sprint(userObj.ID))

	if err != nil {
		log.Println("Failed fetch balance:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	balance := entity.Balance{
		Current:   user.Balance,
		Withdrawn: user.Withdrawn,
	}

	b, err := json.Marshal(balance)
	if err != nil {
		log.Println("Failed marshalling balance:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
