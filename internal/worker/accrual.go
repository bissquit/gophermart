package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/bissquit/gophermart/internal/accrual"
	"github.com/bissquit/gophermart/internal/repository"
)

type AccrualWorker struct {
	storage       repository.GophermartRepository
	accrualClient *accrual.Client
	logger        *slog.Logger
}

func NewAccrualWorker(storage repository.GophermartRepository, accrualClient *accrual.Client, logger *slog.Logger) *AccrualWorker {
	return &AccrualWorker{
		storage:       storage,
		accrualClient: accrualClient,
		logger:        logger,
	}
}

func (w *AccrualWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second) // каждые 5 секунд
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processOrders(ctx)
		}
	}
}

func (w *AccrualWorker) processOrders(ctx context.Context) {
	orders, err := w.storage.GetPendingOrders()
	if err != nil {
		w.logger.Error("failed to get pending orders", "err", err)
		return
	}

	for _, order := range orders {
		resp, err := w.accrualClient.GetOrder(ctx, order.OrderNumber)
		if err != nil {
			w.logger.Error("failed to get order from accrual", "order", order.OrderNumber, "err", err)
			time.Sleep(time.Second)
			continue
		}

		if resp == nil {
			continue
		}

		err = w.storage.UpdateOrderStatus(order.OrderNumber, resp.Status, resp.Accrual)
		if err != nil {
			w.logger.Error("failed to update order", "order", order.OrderNumber, "err", err)
		}

		time.Sleep(time.Second)
	}
}
