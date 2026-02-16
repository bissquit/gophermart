package repository

import (
	"errors"
	"time"
)

var (
	ErrUserAlreadyExists                = errors.New("user already exists")
	ErrUserNotFound                     = errors.New("user not found")
	ErrOrderAlreadyCreatedByUser        = errors.New("order already created by user")
	ErrOrderAlreadyCreatedByAnotherUser = errors.New("order already created by another user")
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
	Accrual     string
	UploadedAt  time.Time
}

type GophermartRepository interface {
	// auth
	CreateUser(login, passwordHash string) (userID string, err error)
	GetUserByLogin(login string) (user User, err error)

	// orders
	CreateOrder(UserID, orderNumber string) error
	//GetOrder(orderNumber string) (order *Order, err error)
	//GetOrdersByUser(login string) (orders []*Order, err error)
}
