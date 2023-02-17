package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/bbt-t/ya-go-d/internal/entity"
)

func (g GophermartHandler) wdAll(w http.ResponseWriter, r *http.Request) {
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

	withdrawals, err := g.s.WithdrawAll(r.Context(), userObj)
	if err != nil {
		log.Println("Can't get withdrawal history:", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		http.Error(w, "no content", http.StatusNoContent)
		return
	}

	b, err := json.Marshal(withdrawals)
	if err != nil {
		log.Println("Can't marshal withdrawal history:", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}