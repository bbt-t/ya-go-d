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

type Queue struct {
	Orders []entity.Order
	*sync.RWMutex
}

func NewQueue() *Queue {
	return &Queue{
		RWMutex: &sync.RWMutex{},
	}
}

func (q *Queue) Push(orders []entity.Order) error {
	q.Lock()
	defer q.Unlock()
	q.Orders = append(orders, q.Orders...)
	return nil
}

func (q *Queue) PushBack(order entity.Order) error {
	q.Lock()
	defer q.Unlock()
	q.Orders = append(q.Orders, order)
	return nil
}

func (q *Queue) GetOrder() (entity.Order, error) {
	q.RLock()
	defer q.RUnlock()
	if len(q.Orders) > 0 {
		order := q.Orders[0]
		q.Orders = q.Orders[1:]
		return order, nil
	}
	return entity.Order{}, ErrEmptyQueue
}
