package handler

import (
	"log"
	"log/slog"
	"mime"
	"net/http"

	"github.com/bissquit/gophermart/internal/repository"
)

type Handlers struct {
	storage   repository.GophermartRepository
	logger    *slog.Logger
	jwtSecret []byte
}

func NewHandlers(storage repository.GophermartRepository, logger *slog.Logger, jwtSecret []byte) *Handlers {
	return &Handlers{
		storage:   storage,
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

func BadRequest(w http.ResponseWriter, message string) {
	log.Printf("bad request: %s", message)
	http.Error(w, message, http.StatusBadRequest) // 400
}

func (h *Handlers) validateContentTypeJSON(w http.ResponseWriter, r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		BadRequest(w, "wrong Content-Type")
		return false
	}
	if mediaType != "application/json" {
		BadRequest(w, "Content-Type must be application/json")
		return false
	}
	return true
}
