package storage

import (
	"sync"

	"github.com/bbt-t/ya-go-d/internal/entity"
)

type UseQueue interface {
	Push(orders []entity.Order) error
	PushBack(order entity.Order) error
	GetOrder() (entity.Order, error)
}

type queue struct {
	orders []entity.Order
	*sync.RWMutex
}

func newQueue() *queue {
	return &queue{
		RWMutex: &sync.RWMutex{},
	}
}

func (q *queue) Push(orders []entity.Order) error {
	q.Lock()
	defer q.Unlock()
	q.orders = append(orders, q.orders...)
	return nil
}

func (q *queue) PushBack(order entity.Order) error {
	q.Lock()
	defer q.Unlock()
	q.orders = append(q.orders, order)
	return nil
}

func (q *queue) GetOrder() (entity.Order, error) {
	q.RLock()
	defer q.RUnlock()
	if len(q.orders) > 0 {
		order := q.orders[0]
		q.orders = q.orders[1:]
		return order, nil
	}
	return entity.Order{}, ErrEmptyQueue
}
