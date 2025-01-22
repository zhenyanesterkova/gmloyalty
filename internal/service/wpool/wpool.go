package wpool

import (
	"context"
	"sync"
	"time"

	"github.com/zhenyanesterkova/gmloyalty/internal/myclient"
	"github.com/zhenyanesterkova/gmloyalty/internal/repository"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/logger"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/order"
)

const (
	StatusNewAccrual        = "REGISTERED"
	StatusProcessingAccrual = "PROCESSING"
	StatusInvalidAccrual    = "INVALID"
	StatusProcessedAccrual  = "PROCESSED"
	SizeQueue               = 1024
)

type WorkerPool struct {
	repo         repository.Store
	Queue        chan order.Order
	TimeOut      chan time.Duration
	logger       logger.LogrusLogger
	accrual      *myclient.AccrualStruct
	errorCh      chan error
	wg           sync.WaitGroup
	countWorkers int
}

func New(
	repo repository.Store,
	logger logger.LogrusLogger,
	accrual *myclient.AccrualStruct,
	countWorkersInPool int,
) *WorkerPool {
	return &WorkerPool{
		Queue:        make(chan order.Order, SizeQueue),
		countWorkers: countWorkersInPool,
		wg:           sync.WaitGroup{},
		logger:       logger,
		accrual:      accrual,
		repo:         repo,
		errorCh:      make(chan error),
	}
}

func (pool *WorkerPool) worker(queue chan order.Order) {
	log := pool.logger.LogrusLog

	for orderInst := range queue {
		ordeAccrualrData, err := pool.accrual.GetOrderInfo(orderInst.Number)
		if err != nil {
			log.Errorf("failed get points from accrual: %v", err)
			queue <- orderInst
			continue
		}

		if ordeAccrualrData.Status == StatusNewAccrual ||
			ordeAccrualrData.Status == StatusProcessingAccrual {
			queue <- orderInst
			continue
		}

		orderInst.Accrual = ordeAccrualrData.Accrual
		orderInst.Status = ordeAccrualrData.Status

		err = pool.repo.ProcessingOrder(context.TODO(), orderInst)
		if err != nil {
			log.Errorf("failed processing order: %v", err)
			queue <- orderInst
			continue
		}
	}
}

func (pool *WorkerPool) Start(ctx context.Context) {
	log := pool.logger.LogrusLog

	log.Info("Starting pool of workers")
	for i := 1; i <= pool.countWorkers; i++ {
		go pool.worker(pool.Queue)
		pool.wg.Add(1)
	}

	pool.wg.Wait()
	log.Info("Stoping pool of workers")
}
