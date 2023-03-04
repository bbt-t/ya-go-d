package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/bbt-t/ya-go-d/internal/entity"
)

func (g GopherMartHandler) wdAll(w http.ResponseWriter, r *http.Request) {
	userObj, ok := r.Context().Value(entity.CtxUserKey("user_id")).(entity.User)
	if !ok {
		log.Println("Wrong value type in context")
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	withdrawals, err := g.s.WithdrawAll(r.Context(), userObj)
	if err != nil {
		log.Printf("Can't get withdrawal history: %+v\n", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		http.Error(w, "no content", http.StatusNoContent)
		return
	}

	b, err := json.Marshal(withdrawals)
	if err != nil {
		log.Printf("Can't marshal withdrawal history: %+v\n", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
