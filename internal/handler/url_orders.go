package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bissquit/gophermart/internal/auth/jwt"
	"github.com/bissquit/gophermart/internal/luhn"
	"github.com/bissquit/gophermart/internal/repository"
)

func (h *Handlers) CreateOrder(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	defer r.Body.Close()
	if err != nil {
		BadRequest(w, "wrong Content-Type")
		return
	}
	if mediaType != "text/plain" {
		BadRequest(w, "Content-Type must be text/plain")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		BadRequest(w, "Cannot read request body")
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		http.Error(w, "order number required", http.StatusBadRequest)
		return
	}

	if _, err := strconv.Atoi(orderNumber); err != nil {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	if !luhn.Valid(orderNumber) {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	userID := r.Context().Value(jwt.UserIDKey).(string)

	err = h.storage.CreateOrder(userID, orderNumber)
	if err == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if errors.Is(err, repository.ErrOrderAlreadyCreatedByUser) {
		w.WriteHeader(http.StatusOK)
		return
	}

	if errors.Is(err, repository.ErrOrderAlreadyCreatedByAnotherUser) {
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	h.logger.Error("create order error", "err", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

type order struct {
	OrderNumber string    `json:"number"`
	Status      string    `json:"status"`
	Accrual     *int      `json:"accrual,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

func (h *Handlers) GetOrdersByUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(jwt.UserIDKey).(string)

	var repoOrders []repository.Order
	repoOrders, err := h.storage.GetOrdersByUser(userID)
	if err != nil {
		h.logger.Error("get orders by user error", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(repoOrders) == 0 {
		w.WriteHeader(http.StatusNoContent) // 204
		return
	}

	userOrders := make([]order, 0, len(repoOrders))
	for _, repoOrder := range repoOrders {
		userOrder := order{
			OrderNumber: repoOrder.OrderNumber,
			Status:      repoOrder.Status,
			Accrual:     repoOrder.Accrual,
			UploadedAt:  repoOrder.UploadedAt,
		}
		userOrders = append(userOrders, userOrder)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(userOrders); err != nil {
		h.logger.Error("error encoding orders", "err", err)
	}
}
