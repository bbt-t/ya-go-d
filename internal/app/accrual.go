package app

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/bbt-t/ya-go-d/internal/adapter/storage"
	"github.com/bbt-t/ya-go-d/internal/app/accrualservice"
	"github.com/bbt-t/ya-go-d/internal/config"
	"github.com/bbt-t/ya-go-d/internal/entity"
)

type workerPool struct {
	jobs    chan entity.Order
	accrual accrualservice.AccrualSystem
	storage storage.DatabaseRepository
}

func newWorkerPool(ctx context.Context, cfg *config.Config, s storage.DatabaseRepository, accrual accrualservice.AccrualSystem) {
	pool := workerPool{
		jobs:    make(chan entity.Order),
		storage: s,
		accrual: accrual,
	}
	for i := 0; i < cfg.Workers; i++ {
		pool.start()
	}
	for {
		job, err := s.GetOrderForUpdate()

		if errors.Is(err, storage.ErrEmptyQueue) {
			time.Sleep(time.Second)
			continue
		}
		if err != nil {
			log.Println("Failed get order for update")
			return
		}

		select {
		case pool.jobs <- job:
			continue
		case <-ctx.Done():
			return
		}
	}

}

func (w *workerPool) start() {
	go func() {
		for {
			work := <-w.jobs

			newOrderInfo, timeToSleep, err := w.accrual.GetOrderUpdates(work)
			if err != nil {
				log.Printf("Failed get update order info: %+v\n", err)
				err := w.storage.Push([]entity.Order{work})
				if err != nil {
					log.Printf("Failed push order in queue: %+v\n", err)
				}
				if timeToSleep > 0 {
					time.Sleep(time.Duration(timeToSleep) * time.Second)
				}
				continue
			}

			if newOrderInfo.Status != work.Status {
				work.Accrual, work.Status = newOrderInfo.Accrual, newOrderInfo.Status
				if err := w.storage.UpdateOrders(context.Background(), work); err != nil {
					log.Printf("Failed update order: %+v\n", err)
				}
			} else {
				if err := w.storage.PushBack(work); err != nil {
					log.Printf("Failed push order in queue: %+v\n", err)
				}
			}
		}
	}()
}
