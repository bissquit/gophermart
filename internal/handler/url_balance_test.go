package handler

import (
	"bytes"
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
	"github.com/stretchr/testify/require"
)

func Test_GetUserBalance(t *testing.T) {
	var jwtSecret = []byte("your-secret-key-min-32-chars-long!")

	tests := []struct {
		name    string
		userID  string
		storage *fakeStorage
		code    int
	}{
		{
			name:   "success",
			userID: "fake-user-uuid",
			storage: &fakeStorage{
				current:   55.3,
				withdrawn: 320.0,
			},
			code: http.StatusOK,
		},
		{
			name:   "emulate error getting user balance",
			userID: "fake-user-uuid",
			storage: &fakeStorage{
				getUserBalanceErr: errors.New("fake-get-user-balance-error"),
			},
			code: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))

			r := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)

			ctx := context.WithValue(r.Context(), jwt.UserIDKey, tt.userID)
			r = r.WithContext(ctx)

			w := httptest.NewRecorder()

			h := NewHandlers(tt.storage, logger, jwtSecret)
			h.GetUserBalance(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.code, res.StatusCode)
		})
	}
}

func Test_RequestUserWithdrawal(t *testing.T) {
	var jwtSecret = []byte("your-secret-key-min-32-chars-long!")
	type input struct {
		userID               string
		body                 *userWithdrawal
		emulateIncorrectJSON bool
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
			name: "successful withdrawal",
			input: input{
				userID: "fake-user-uuid",
				body: &userWithdrawal{
					OrderNumber: "17893729974",
					Sum:         55.3,
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
			name: "corrupted body",
			input: input{
				body: &userWithdrawal{
					OrderNumber: "17893729974",
					Sum:         55.3,
				},
				emulateIncorrectJSON: true,
				storage:              newFakeStorage(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "incorrect body",
			input: input{
				userID:               "fake-user-uuid",
				body:                 &userWithdrawal{},
				contentType:          "application/json",
				emulateIncorrectBody: true,
				storage:              newFakeStorage(),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "emulate ErrLowBalance",
			input: input{
				body: &userWithdrawal{
					OrderNumber: "17893729974",
					Sum:         55.3,
				},
				contentType: "application/json",
				storage: &fakeStorage{
					requestUserWithdrawalErr: repository.ErrLowBalance,
				},
			},
			want: want{
				code: http.StatusPaymentRequired,
			},
		},
		{
			name: "emulate ErrBalanceOrderAlreadyWithdrawn",
			input: input{
				body: &userWithdrawal{
					OrderNumber: "17893729974",
					Sum:         55.3,
				},
				contentType: "application/json",
				storage: &fakeStorage{
					requestUserWithdrawalErr: repository.ErrBalanceOrderAlreadyWithdrawn,
				},
			},
			want: want{
				code: http.StatusUnprocessableEntity,
			},
		},
		{
			name: "emulate unknown storage error",
			input: input{
				body: &userWithdrawal{
					OrderNumber: "17893729974",
					Sum:         55.3,
				},
				contentType: "application/json",
				storage: &fakeStorage{
					requestUserWithdrawalErr: errors.New("fake-get-user-withdrawal-error"),
				},
			},
			want: want{
				code: http.StatusInternalServerError,
			},
		},
		{
			name: "incorrect lhun",
			input: input{
				body: &userWithdrawal{
					OrderNumber: "123",
					Sum:         55.3,
				},
				contentType: "application/json",
				storage:     &fakeStorage{},
			},
			want: want{
				code: http.StatusUnprocessableEntity,
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
				body = strings.NewReader(tt.input.body.OrderNumber)
			} else {
				// valid JSON body
				req := tt.input.body
				b, err := json.Marshal(req)
				require.NoError(t, err)
				body = bytes.NewReader(b)
			}
			if tt.input.emulateIncorrectBody {
				body = strings.NewReader("")
			}

			r := httptest.NewRequest(http.MethodGet, "/api/user/balance/withdraw", body)
			if tt.input.contentType != "" {
				r.Header.Set("Content-Type", tt.input.contentType)
			}

			ctx := context.WithValue(r.Context(), jwt.UserIDKey, tt.input.userID)
			r = r.WithContext(ctx)

			w := httptest.NewRecorder()

			h := NewHandlers(tt.input.storage, logger, jwtSecret)
			h.RequestUserWithdrawal(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}

func Test_GetUserWithdrawals(t *testing.T) {
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
			name: "successful get withdrawals",
			input: input{
				userID: "fake-user-uuid",
				storage: &fakeStorage{
					withdrawals: []repository.Withdrawal{
						{
							ID:          "withdrawal-1",
							UserID:      "fake-user-uuid",
							OrderNumber: "17893729974",
							Sum:         100.5,
							ProcessedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
						},
						{
							ID:          "withdrawal-2",
							UserID:      "fake-user-uuid",
							OrderNumber: "12345678903",
							Sum:         50.0,
							ProcessedAt: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
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
			name: "no withdrawals (empty list)",
			input: input{
				userID: "fake-user-uuid",
				storage: &fakeStorage{
					withdrawals: []repository.Withdrawal{},
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
					getUserWithdrawalsErr: errors.New("fake-storage-error"),
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

			r := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)

			ctx := context.WithValue(r.Context(), jwt.UserIDKey, tt.input.userID)
			r = r.WithContext(ctx)

			w := httptest.NewRecorder()

			h := NewHandlers(tt.input.storage, logger, jwtSecret)
			h.GetUserWithdrawals(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)

			if tt.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)
			}

			if tt.want.bodyEmpty {
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Empty(t, body)
			}

			// Дополнительная проверка для успешного случая
			if tt.want.code == http.StatusOK {
				var result []withdrawal
				err := json.NewDecoder(res.Body).Decode(&result)
				require.NoError(t, err)
				assert.Len(t, result, 2)
				assert.Equal(t, "17893729974", result[0].Order)
				assert.Equal(t, 100.5, result[0].Sum)
			}
		})
	}
}
