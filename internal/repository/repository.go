package repository

import (
	"errors"
	"time"
)

var (
	// users
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	// orders
	ErrOrderAlreadyCreatedByUser        = errors.New("order already created by user")
	ErrOrderAlreadyCreatedByAnotherUser = errors.New("order already created by another user")
	// balance
)

type User struct {
	ID           string
	Login        string
	PasswordHash string
}

type Order struct {
	ID          string
	UserID      string
	OrderNumber string
	Status      string
	Accrual     *int // *int may be nil
	UploadedAt  time.Time
}

type Withdrawal struct {
	ID          string
	UserID      string
	OrderNumber string
	Sum         int
	ProcessedAt *time.Time
}

type GophermartRepository interface {
	// auth
	CreateUser(login, passwordHash string) (userID string, err error)
	GetUserByLogin(login string) (user User, err error)

	// orders
	CreateUserOrder(UserID, orderNumber string) error
	GetUserOrders(UserID string) (orders []Order, err error)

	// balance
	//GetUserBalance(userID string) (balance int, err error)
	//WithdrawUserBalance(userID string, amount int) (balance int, err error)
	//GetUserWithdrawals(userID string)
}
