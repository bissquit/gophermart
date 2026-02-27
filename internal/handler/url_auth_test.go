package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bissquit/gophermart/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Register(t *testing.T) {
	var jwtSecret = []byte("your-secret-key-min-32-chars-long!")
	type input struct {
		body                 user
		emulateIncorrectJSON bool
		contentType          string
		storage              *fakeStorage
	}
	type want struct {
		code        int
		contentType string
	}

	tests := []struct {
		name  string
		input input
		want  want
	}{
		{
			name: "successful registration",
			input: input{
				body: user{
					UserID:   "fake-user-uuid",
					Login:    "fake-user-uuid",
					Password: "fake-password",
				},
				contentType: "application/json",
				storage:     newFakeStorage(),
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name: "wrong content type",
			input: input{
				contentType: "text/plain",
				storage:     newFakeStorage(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "dont set content type header",
			input: input{
				storage: newFakeStorage(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "empty body",
			input: input{
				emulateIncorrectJSON: true,
				contentType:          "application/json",
				storage:              newFakeStorage(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "empty username",
			input: input{
				body: user{
					UserID:   "fake-user-uuid",
					Login:    "",
					Password: "fake-password",
				},
				contentType: "application/json",
				storage:     newFakeStorage(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "fake password (emulate bcrypt error)",
			input: input{
				body: user{
					UserID: "fake-user-uuid",
					Login:  "fake-user-uuid",
					// password.Hash(userItem.Password) returns error when
					// password is longer than 73 bytes
					Password: strings.Repeat("a", 73),
				},
				contentType: "application/json",
				storage:     newFakeStorage(),
			},
			want: want{
				code: http.StatusInternalServerError,
			},
		},
		{
			name: "emulate ErrUserAlreadyExists",
			input: input{
				body: user{
					UserID:   "fake-user-uuid",
					Login:    "fake-user-uuid",
					Password: "fake-password",
				},
				contentType: "application/json",
				storage: &fakeStorage{
					createUserErr: repository.ErrUserAlreadyExists,
				},
			},
			want: want{
				code: http.StatusConflict,
			},
		},
		{
			name: "emulate unknown storage error",
			input: input{
				body: user{
					UserID:   "fake-user-uuid",
					Login:    "fake-user-uuid",
					Password: "fake-password",
				},
				contentType: "application/json",
				storage: &fakeStorage{
					createUserErr: errors.New("fake error"),
				},
			},
			want: want{
				code: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))

			var body io.Reader
			if tt.input.emulateIncorrectJSON {
				// raw body: intentionally broken JSON or empty string, send as-is
				body = strings.NewReader(tt.input.body.Login)
			} else {
				// valid JSON body
				req := tt.input.body
				b, err := json.Marshal(req)
				require.NoError(t, err)
				body = bytes.NewReader(b)
			}

			// create request
			r := httptest.NewRequest(http.MethodPost, "/api/user/register", body)
			if tt.input.contentType != "" {
				r.Header.Set("Content-Type", tt.input.contentType)
			}
			// create ResponseWriter
			w := httptest.NewRecorder()

			// create handlers
			h := NewHandlers(tt.input.storage, logger, jwtSecret)
			h.Register(w, r)

			// get result
			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			t.Log(string(resBody))
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}

func Test_Login(t *testing.T) {
	var jwtSecret = []byte("your-secret-key-min-32-chars-long!")
	type input struct {
		body                 user
		emulateIncorrectJSON bool
		contentType          string
		storage              *fakeStorage
	}
	type want struct {
		code        int
		contentType string
	}

	tests := []struct {
		name  string
		input input
		want  want
	}{
		{
			name: "successful login",
			input: input{
				body: user{
					UserID:   "fake-user-uuid",
					Login:    "fake-user-uuid",
					Password: "fake-password",
				},
				contentType: "application/json",
				storage:     newFakeStorage(),
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name: "wrong password",
			input: input{
				body: user{
					UserID:   "fake-user-uuid",
					Login:    "fake-user-uuid",
					Password: "wrong-fake-password",
				},
				contentType: "application/json",
				storage:     newFakeStorage(),
			},
			want: want{
				code: http.StatusUnauthorized,
			},
		},
		{
			name: "wrong content type",
			input: input{
				contentType: "text/plain",
				storage:     newFakeStorage(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "dont set content type header",
			input: input{
				storage: newFakeStorage(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "empty body",
			input: input{
				emulateIncorrectJSON: true,
				contentType:          "application/json",
				storage:              newFakeStorage(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "empty username",
			input: input{
				body: user{
					UserID:   "fake-user-uuid",
					Login:    "",
					Password: "fake-password",
				},
				contentType: "application/json",
				storage:     newFakeStorage(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "emulate ErrUserNotFound",
			input: input{
				body: user{
					UserID:   "fake-user-uuid",
					Login:    "fake-user-uuid",
					Password: "fake-password",
				},
				contentType: "application/json",
				storage: &fakeStorage{
					getUserByLoginErr: repository.ErrUserNotFound,
				},
			},
			want: want{
				code: http.StatusUnauthorized,
			},
		},
		{
			name: "emulate unknown storage error",
			input: input{
				body: user{
					UserID:   "fake-user-uuid",
					Login:    "fake-user-uuid",
					Password: "fake-password",
				},
				contentType: "application/json",
				storage: &fakeStorage{
					getUserByLoginErr: errors.New("fake error"),
				},
			},
			want: want{
				code: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))

			var body io.Reader
			if tt.input.emulateIncorrectJSON {
				// raw body: intentionally broken JSON or empty string, send as-is
				body = strings.NewReader(tt.input.body.Login)
			} else {
				// valid JSON body
				req := tt.input.body
				b, err := json.Marshal(req)
				require.NoError(t, err)
				body = bytes.NewReader(b)
			}

			// create request
			r := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
			if tt.input.contentType != "" {
				r.Header.Set("Content-Type", tt.input.contentType)
			}
			// create ResponseWriter
			w := httptest.NewRecorder()

			// create handlers
			h := NewHandlers(tt.input.storage, logger, jwtSecret)
			h.Login(w, r)

			// get result
			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			t.Log(string(resBody))
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}
