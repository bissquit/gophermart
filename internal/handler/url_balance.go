package handler

import (
	"encoding/json"
	"net/http"

	"github.com/bissquit/gophermart/internal/auth/jwt"
)

type userBalance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (h *Handlers) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(jwt.UserIDKey).(string)

	current, withdrawn, err := h.storage.GetUserBalance(userID)
	if err != nil {
		h.logger.Error("error getting user balance", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	balance := userBalance{
		Current:   current,
		Withdrawn: withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		h.logger.Error("error encoding balance", "err", err)
	}
}
