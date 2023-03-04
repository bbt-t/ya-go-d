package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bbt-t/ya-go-d/internal/entity"
	"github.com/bbt-t/ya-go-d/pkg"
)

func (g GopherMartHandler) wdAll(w http.ResponseWriter, r *http.Request) {
	userObj, ok := r.Context().Value(entity.CtxUserKey("user_id")).(entity.User)
	if !ok {
		pkg.Log.Info("Wrong value type in context")
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	withdrawals, err := g.s.WithdrawAll(r.Context(), userObj)
	if err != nil {
		pkg.Log.Warn(err.Error())
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		http.Error(w, "no content", http.StatusNoContent)
		return
	}

	b, err := json.Marshal(withdrawals)
	if err != nil {
		pkg.Log.Err(err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
