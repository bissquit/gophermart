package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/bissquit/gophermart/internal/auth/jwt"
	"github.com/bissquit/gophermart/internal/luhn"
	"github.com/bissquit/gophermart/internal/repository"
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError) // 500
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

type userWithdrawal struct {
	OrderNumber string  `json:"order"`
	Sum         float64 `json:"sum"`
}

func (h *Handlers) RequestUserWithdrawal(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if !h.validateContentTypeJSON(w, r) {
		return
	}

	userID := r.Context().Value(jwt.UserIDKey).(string)

	var userWithdrawalItem *userWithdrawal
	if err := json.NewDecoder(r.Body).Decode(&userWithdrawalItem); err != nil {
		h.logger.Error("decode user withdrawal error", "err", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest) // 400
		return
	}

	if !luhn.Valid(userWithdrawalItem.OrderNumber) {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity) // 422
		return
	}

	err := h.storage.RequestUserWithdrawal(userID, userWithdrawalItem.OrderNumber, userWithdrawalItem.Sum)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
		return
	case errors.Is(err, repository.ErrLowBalance):
		http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired) // 402
		return
	case errors.Is(err, repository.ErrBalanceOrderAlreadyWithdrawn):
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity) // 422
		return
	}

	h.logger.Error("error requesting user withdrawal", "err", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError) // 500
}

type withdrawal struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (h *Handlers) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(jwt.UserIDKey).(string)

	withdrawals, err := h.storage.GetUserWithdrawals(userID)
	if err != nil {
		h.logger.Error("get user withdrawals error", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError) // 500
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent) // 204
		return
	}

	result := make([]withdrawal, 0, len(withdrawals))
	for _, wdrl := range withdrawals {
		result = append(result, withdrawal{
			Order:       wdrl.OrderNumber,
			Sum:         wdrl.Sum,
			ProcessedAt: wdrl.ProcessedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Error("error encoding withdrawals", "err", err)
	}
}
