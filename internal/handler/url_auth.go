package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bissquit/gophermart/internal/auth/jwt"
	"github.com/bissquit/gophermart/internal/password"
	"github.com/bissquit/gophermart/internal/repository"
)

type user struct {
	UserID   string `json:"user_id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	// check request headers
	defer r.Body.Close()
	if !h.validateContentTypeJSON(w, r) {
		return
	}

	// get login/password
	var userItem *user
	if err := json.NewDecoder(r.Body).Decode(&userItem); err != nil {
		h.logger.Error("decode user error", "err", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// fail fast
	if userItem.Login == "" || userItem.Password == "" {
		http.Error(w, "login and password required", http.StatusBadRequest)
		return
	}

	// get hash
	passwordHash, err := password.Hash(userItem.Password)
	if err != nil {
		h.logger.Error("error while hashing password", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// create user
	var userID string
	userID, err = h.storage.CreateUser(userItem.Login, passwordHash)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		h.logger.Error("create user error", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// generate token
	token, err := jwt.GenerateToken(userID, userItem.Login, h.jwtSecret)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		h.logger.Error("error encoding token", "error", err)
		// do nothing, 'http.Error()' is useless because headers are already sent
		return
	}
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	// check request headers
	defer r.Body.Close()
	if !h.validateContentTypeJSON(w, r) {
		return
	}

	// get login/password
	var userItem *user
	if err := json.NewDecoder(r.Body).Decode(&userItem); err != nil {
		h.logger.Error("decode user error", "err", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// fail fast
	if userItem.Login == "" || userItem.Password == "" {
		http.Error(w, "login and password required", http.StatusBadRequest)
		return
	}

	// get user from db with same login
	var u repository.User
	u, err := h.storage.GetUserByLogin(userItem.Login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		h.logger.Error("get user error", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// check hash
	if !password.CheckHash(userItem.Password, u.PasswordHash) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// generate token
	token, err := jwt.GenerateToken(u.ID, u.Login, h.jwtSecret)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		h.logger.Error("error encoding token", "error", err)
		// do nothing, 'http.Error()' is useless because headers are already sent
		return
	}
}
