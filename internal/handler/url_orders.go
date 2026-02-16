package handler

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

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
