package usecase

import (
	"context"

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

type GophermartService struct {
	repo DatabaseRepository
}

func NewGophermart(r DatabaseRepository) *GophermartService {
	return &GophermartService{
		repo: r,
	}
}

func (s *GophermartService) NewUser(ctx context.Context, user entity.User) (int, error) {
	return s.repo.NewUser(ctx, user)
}

func (s *GophermartService) GetUser(ctx context.Context, search, value string) (entity.User, error) {
	return s.repo.GetUser(ctx, search, value)
}

func (s *GophermartService) Withdraw(ctx context.Context, user entity.User, wd entity.Withdraw) error {
	return s.repo.Withdraw(ctx, user, wd)
}

func (s *GophermartService) WithdrawAll(ctx context.Context, user entity.User) ([]entity.Withdraw, error) {
	return s.repo.WithdrawAll(ctx, user)
}

func (s *GophermartService) AddOrder(ctx context.Context, order entity.Order) error {
	return s.repo.AddOrder(ctx, order)
}

func (s *GophermartService) OrdersAll(ctx context.Context, user entity.User) ([]entity.Order, error) {
	return s.repo.OrdersAll(ctx, user)
}

func (s *GophermartService) GetOrdersForUpdate(ctx context.Context) ([]entity.Order, error) {
	return s.repo.GetOrdersForUpdate(ctx)
}

func (s *GophermartService) GetOrderForUpdate() (entity.Order, error) {
	return s.repo.GetOrderForUpdate()
}

func (s *GophermartService) UpdateOrders(ctx context.Context, orders ...entity.Order) error {
	return s.repo.UpdateOrders(ctx, orders...)
}

func (s *GophermartService) Push(orders []entity.Order) error {
	return s.repo.Push(orders)
}

func (s *GophermartService) PushBack(order entity.Order) error {
	return s.repo.PushBack(order)
}
