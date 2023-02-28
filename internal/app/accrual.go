package app

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/bbt-t/ya-go-d/internal/adapter/storage"
	"github.com/bbt-t/ya-go-d/internal/app/accrualservice"
	"github.com/bbt-t/ya-go-d/internal/config"
	"github.com/bbt-t/ya-go-d/internal/entity"
)

type timer struct {
	Time *time.Timer
	*sync.RWMutex
}

type workerPool struct {
	jobs    chan entity.Order
	accrual accrualservice.AccrualSystem
	storage storage.DatabaseRepository
	timer   timer
}

func newWorkerPool(ctx context.Context, cfg *config.Config, s storage.DatabaseRepository, accrual accrualservice.AccrualSystem) {
	pool := workerPool{
		jobs:    make(chan entity.Order),
		storage: s,
		accrual: accrual,
		timer: timer{
			RWMutex: &sync.RWMutex{},
		},
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
			log.Printf("Sent job to worker: %v", job)
		case <-ctx.Done():
			log.Println("Shutdown")
			return
		}
	}

}

func (w *workerPool) start() {
	go func() {
		for {
			work := <-w.jobs

			newOrderInfo, _, err := w.accrual.GetOrderUpdates(work)
			if err != nil {
				log.Printf("Failed get update order info: %+v\n", err)
				err := w.storage.Push([]entity.Order{work})
				if err != nil {
					log.Printf("Failed push order in queue: %+v\n", err)
				}
				//if sleep > 0 {
				//	w.timer.Lock()
				//	w.timer.Time = time.NewTimer(time.Duration(sleep) * time.Second)
				//	w.timer.Unlock()
				//}
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
