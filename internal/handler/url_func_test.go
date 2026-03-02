package handler

import (
	"github.com/bissquit/gophermart/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type fakeStorage struct {
	createUserErr            error
	getUserByLoginErr        error
	createUserOrderErr       error
	getUserOrdersErr         error
	getUserBalanceErr        error
	requestUserWithdrawalErr error
	getUserWithdrawalsErr    error
	getPendingOrdersErr      error
	updateOrderStatusErr     error

	// GetUserBalance
	current, withdrawn float64

	// GetUserWithdrawals
	withdrawals []repository.Withdrawal

	// GetUserOrders
	orders []repository.Order
}

func newFakeStorage() *fakeStorage { return &fakeStorage{} }

func (f *fakeStorage) CreateUser(login, passwordHash string) (userID string, err error) {
	if f.createUserErr != nil {
		return "", f.createUserErr
	}
	return "fake-user-uuid", nil
}

func (f *fakeStorage) GetUserByLogin(login string) (user repository.User, err error) {
	if f.getUserByLoginErr != nil {
		return repository.User{}, f.getUserByLoginErr
	}

	realHash, _ := bcrypt.GenerateFromPassword([]byte("fake-password"), bcrypt.DefaultCost)

	return repository.User{
		ID:           "fake-user-uuid",
		Login:        "fake-user-uuid",
		PasswordHash: string(realHash),
	}, nil
}

func (f *fakeStorage) CreateUserOrder(UserID, orderNumber string) error {
	if f.createUserOrderErr != nil {
		return f.createUserOrderErr
	}
	return nil
}

func (f *fakeStorage) GetUserOrders(UserID string) (orders []repository.Order, err error) {
	if f.getUserOrdersErr != nil {
		return nil, f.getUserOrdersErr
	}
	if f.orders != nil {
		return f.orders, nil
	}
	return []repository.Order{}, nil
}

func (f *fakeStorage) GetUserBalance(userID string) (current, withdrawn float64, err error) {
	if f.getUserBalanceErr != nil {
		return f.current, f.withdrawn, f.getUserBalanceErr
	}
	return 0.0, 0.0, nil
}

func (f *fakeStorage) RequestUserWithdrawal(userID string, orderNumber string, sum float64) error {
	if f.requestUserWithdrawalErr != nil {
		return f.requestUserWithdrawalErr
	}
	return nil
}

func (f *fakeStorage) GetUserWithdrawals(userID string) ([]repository.Withdrawal, error) {
	if f.getUserWithdrawalsErr != nil {
		return nil, f.getUserWithdrawalsErr
	}
	if f.withdrawals != nil {
		return f.withdrawals, nil
	}
	return []repository.Withdrawal{}, nil
}

func (f *fakeStorage) GetPendingOrders() ([]repository.Order, error) {
	if f.getPendingOrdersErr != nil {
		return nil, f.getPendingOrdersErr
	}
	return make([]repository.Order, 0), nil
}

func (f *fakeStorage) UpdateOrderStatus(orderNumber, status string, accrual *float64) error {
	if f.updateOrderStatusErr != nil {
		return f.updateOrderStatusErr
	}
	return nil
}
