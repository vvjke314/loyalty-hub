package accrual

import (
	"context"
	"fmt"
	"time"

	"github.com/vvjke314/itk-courses/loyalityhub/internal/services"
)

type AccrualWorker struct {
	rate    *time.Ticker
	service *services.AccrualWorkerService
}

// фабрика для воркера
func NewAccrualWorker(workerRate time.Duration, orderService *services.AccrualWorkerService) *AccrualWorker {
	return &AccrualWorker{
		rate:    time.NewTicker(workerRate),
		service: orderService,
	}
}

func (worker *AccrualWorker) Work(ctx context.Context, errChan chan<- error) error {
	for {
		select {
		case <-worker.rate.C:
			err := worker.doWork(ctx)
			if err != nil {
				errChan <- err
			}
		case <-ctx.Done():
			return fmt.Errorf("accrual worker canceled")
		}
	}
}

func (worker *AccrualWorker) doWork(ctx context.Context) error {
	return worker.service.UpdateOrders(ctx)
}
