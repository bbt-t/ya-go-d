package storage

import (
	"context"
	"github.com/bbt-t/ya-go-d/internal/config"

	"github.com/bbt-t/ya-go-d/internal/entity"
)

type DatabaseRepository interface {
	/*
		Interface for using DB.
	*/
	NewUser(ctx context.Context, user entity.User) (int, error)
	GetUser(ctx context.Context, search, value string) (entity.User, error)
	Withdraw(ctx context.Context, user entity.User, wd entity.Withdraw) error
	WithdrawAll(ctx context.Context, user entity.User) ([]entity.Withdraw, error)
	AddOrder(ctx context.Context, order entity.Order) error
	OrdersAll(ctx context.Context, user entity.User) ([]entity.Order, error)
	GetOrdersForUpdate(ctx context.Context) ([]entity.Order, error)
	GetOrderForUpdate() (entity.Order, error)
	UpdateOrders(ctx context.Context, orders ...entity.Order) error
	Push(orders []entity.Order) error
	PushBack(order entity.Order) error
}

func NewStorage(cfg *config.Config) DatabaseRepository {
	return newDBStorage(cfg)
}
