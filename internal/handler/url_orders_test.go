package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bissquit/gophermart/internal/auth/jwt"
	"github.com/bissquit/gophermart/internal/repository"
	"github.com/stretchr/testify/assert"
)

func Test_CreateUserOrder(t *testing.T) {
	var jwtSecret = []byte("your-secret-key-min-32-chars-long!")
	type input struct {
		userID               string
		body                 string
		emulateIncorrectBody bool
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
			name: "successful order creation",
			input: input{
				userID:      "fake-user-uuid",
				body:        "17893729974",
				contentType: "text/plain",
				storage:     newFakeStorage(),
			},
			want: want{
				code: http.StatusAccepted,
			},
		},
		{
			name: "wrong content type",
			input: input{
				userID:      "fake-user-uuid",
				body:        "17893729974",
				contentType: "application/json",
				storage:     newFakeStorage(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "emulate missing content type",
			input: input{
				userID:  "fake-user-uuid",
				body:    "17893729974",
				storage: newFakeStorage(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "empty order number",
			input: input{
				body:        "",
				contentType: "text/plain",
				storage:     newFakeStorage(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "wrong order number (not-digits)",
			input: input{
				body:        "not-digits",
				contentType: "text/plain",
				storage:     newFakeStorage(),
			},
			want: want{
				code: http.StatusUnprocessableEntity,
			},
		},
		{
			name: "wrong order number (not luhn)",
			input: input{
				body:        "123321",
				contentType: "text/plain",
				storage:     newFakeStorage(),
			},
			want: want{
				code: http.StatusUnprocessableEntity,
			},
		},
		{
			name: "emulate ErrOrderAlreadyCreatedByUser storage error",
			input: input{
				userID:      "fake-user-uuid",
				body:        "17893729974",
				contentType: "text/plain",
				storage: &fakeStorage{
					createUserOrderErr: repository.ErrOrderAlreadyCreatedByUser,
				},
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "emulate ErrOrderAlreadyCreatedByAnotherUser storage error",
			input: input{
				userID:      "fake-user-uuid",
				body:        "17893729974",
				contentType: "text/plain",
				storage: &fakeStorage{
					createUserOrderErr: repository.ErrOrderAlreadyCreatedByAnotherUser,
				},
			},
			want: want{
				code: http.StatusConflict,
			},
		},
		{
			name: "emulate unknown storage error",
			input: input{
				userID:      "fake-user-uuid",
				body:        "17893729974",
				contentType: "text/plain",
				storage: &fakeStorage{
					createUserOrderErr: errors.New("some error"),
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

			body := strings.NewReader(tt.input.body)

			r := httptest.NewRequest(http.MethodPost, "/api/user/orders", body)

			if tt.input.contentType != "" {
				r.Header.Set("Content-Type", tt.input.contentType)
			}

			ctx := context.WithValue(r.Context(), jwt.UserIDKey, tt.input.userID)
			r = r.WithContext(ctx)

			w := httptest.NewRecorder()

			h := NewHandlers(tt.input.storage, logger, jwtSecret)
			h.CreateUserOrder(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}

func Test_GetUserOrders(t *testing.T) {
	var jwtSecret = []byte("your-secret-key-min-32-chars-long!")

	type input struct {
		userID  string
		storage *fakeStorage
	}
	type want struct {
		code        int
		contentType string
		bodyEmpty   bool
	}

	tests := []struct {
		name  string
		input input
		want  want
	}{
		{
			name: "successful get orders",
			input: input{
				userID: "fake-user-uuid",
				storage: &fakeStorage{
					orders: []repository.Order{
						{
							ID:          "order-1",
							UserID:      "fake-user-uuid",
							OrderNumber: "17893729974",
							Status:      "PROCESSED",
							Accrual:     func() *float64 { v := 500.0; return &v }(),
							UploadedAt:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
						},
						{
							ID:          "order-2",
							UserID:      "fake-user-uuid",
							OrderNumber: "12345678903",
							Status:      "PROCESSING",
							Accrual:     nil,
							UploadedAt:  time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name: "no orders (empty list)",
			input: input{
				userID: "fake-user-uuid",
				storage: &fakeStorage{
					orders: []repository.Order{},
				},
			},
			want: want{
				code:      http.StatusNoContent,
				bodyEmpty: true,
			},
		},
		{
			name: "emulate storage error",
			input: input{
				userID: "fake-user-uuid",
				storage: &fakeStorage{
					getUserOrdersErr: errors.New("fake-storage-error"),
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

			r := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)

			ctx := context.WithValue(r.Context(), jwt.UserIDKey, tt.input.userID)
			r = r.WithContext(ctx)

			w := httptest.NewRecorder()

			h := NewHandlers(tt.input.storage, logger, jwtSecret)
			h.GetUserOrders(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)

			if tt.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)
			}

			if tt.want.bodyEmpty {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)
				assert.Empty(t, body)
			}

			// additional checks for successful case
			if tt.want.code == http.StatusOK {
				var result []order
				err := json.NewDecoder(res.Body).Decode(&result)
				assert.NoError(t, err)
				assert.Len(t, result, 2)
				assert.Equal(t, "17893729974", result[0].OrderNumber)
				assert.Equal(t, "PROCESSED", result[0].Status)
				assert.NotNil(t, result[0].Accrual)
				assert.Equal(t, 500.0, *result[0].Accrual)
				assert.Nil(t, result[1].Accrual)
			}
		})
	}
}
