package app

import (
	"context"
	"errors"
	"time"

	"github.com/bbt-t/ya-go-d/internal/adapter/storage"
	"github.com/bbt-t/ya-go-d/internal/app/accrualservice"
	"github.com/bbt-t/ya-go-d/internal/config"
	"github.com/bbt-t/ya-go-d/internal/entity"
	"github.com/bbt-t/ya-go-d/pkg"
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
				pkg.Log.Info(err.Error())
				err := w.storage.Push([]entity.Order{work})
				if err != nil {
					pkg.Log.Info(err.Error())
				}
				if timeToSleep > 0 {
					time.Sleep(time.Duration(timeToSleep) * time.Second)
				}
				continue
			}

			if newOrderInfo.Status != work.Status {
				work.Accrual, work.Status = newOrderInfo.Accrual, newOrderInfo.Status
				if err := w.storage.UpdateOrders(context.Background(), work); err != nil {
					pkg.Log.Info(err.Error())
				}
			} else {
				if err := w.storage.PushBack(work); err != nil {
					pkg.Log.Info(err.Error())
				}
			}
		}
	}()
}
