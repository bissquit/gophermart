package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/bissquit/gophermart/internal/auth/jwt"
	"github.com/bissquit/gophermart/internal/config"
	"github.com/bissquit/gophermart/internal/handler"
	"github.com/bissquit/gophermart/internal/logging"
	"github.com/bissquit/gophermart/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	config  *config.Config
	storage repository.GophermartRepository
	router  *chi.Mux
	DB      *pgxpool.Pool
	logger  *slog.Logger
}

func NewServer(config *config.Config,
	storage repository.GophermartRepository,
	db *pgxpool.Pool,
	logger *slog.Logger) *Server {

	s := &Server{
		config:  config,
		storage: storage,
		router:  chi.NewRouter(),
		DB:      db,
		logger:  logger,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	secret := []byte("your-secret-key-min-32-chars-long!")

	h := handler.NewHandlers(s.storage, s.logger, secret)

	// middlewares
	s.router.Use(logging.Logger(s.logger))

	// public endpoints
	s.router.Post("/api/user/register", h.Register)
	s.router.Post("/api/user/login", h.Login)
	s.router.Get("/ping", s.Ping)

	s.router.Group(func(r chi.Router) {
		r.Use(jwt.JWT(secret))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("It works!\n"))
		})
		r.Post("/api/user/orders", h.CreateOrder)
	})
}

func (s *Server) Ping(w http.ResponseWriter, r *http.Request) {
	pingCtx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := s.DB.Ping(pingCtx); err != nil {
		s.logger.Error("db ping failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) Handler() http.Handler {
	return s.router
}
