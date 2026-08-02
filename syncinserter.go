// Package syncinserter provides batch inserter that merges concurrent insert operations into batch insert.
package syncinserter

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrInvalidMaxBatchSize indicates invalid max batch size parameter in New.
	ErrInvalidMaxBatchSize = errors.New("syncinserter: batch size should be greater 0")
	// ErrInvalidPeriod indicates invalid period parameter in New.
	ErrInvalidPeriod  = errors.New("syncinserter: period should be greater 0")
	notifyChannelPool = sync.Pool{
		New: func() any {
			return make(chan error, 1)
		},
	}
)

type (
	queueElement[Row any] struct {
		notifyCh chan error
		row      Row
	}

	// SyncInserter is bulk inserter with sync Insert API. It merges all Insert calls during period and merges them in one batch call.
	SyncInserter[Row any] struct {
		maxBatchSize    int
		batchInsertFunc func(rows []Row) error
		queue           chan queueElement[Row]
		rowsBuffer      []Row
		notifiesBuffer  []chan error
	}
)

// New initializes SyncInserter.
func New[Row any](ctx context.Context, batchInsertFunc func(rows []Row) error, maxBatchSize int, period time.Duration) (*SyncInserter[Row], error) {
	if maxBatchSize <= 0 {
		return nil, ErrInvalidMaxBatchSize
	}

	if period <= 0 {
		return nil, ErrInvalidPeriod
	}

	syncInserter := &SyncInserter[Row]{
		maxBatchSize:    maxBatchSize,
		batchInsertFunc: batchInsertFunc,
		queue:           make(chan queueElement[Row]),
		rowsBuffer:      make([]Row, 0, maxBatchSize),
		notifiesBuffer:  make([]chan error, 0, maxBatchSize),
	}

	go func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				syncInserter.send()

				return
			case <-ticker.C:
				syncInserter.send()
			}
		}
	}()

	return syncInserter, nil
}

// Insert adds a row to the current batch and blocks until context canceled or the batch is processed.
func (syncInserter *SyncInserter[Row]) Insert(ctx context.Context, row Row) error {
	notifyCh := notifyChannelPool.Get().(chan error)
	defer notifyChannelPool.Put(notifyCh)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case syncInserter.queue <- queueElement[Row]{row: row, notifyCh: notifyCh}:
		return <-notifyCh
	}
}

func (syncInserter *SyncInserter[Row]) send() {
	for empty := false; len(syncInserter.rowsBuffer) < cap(syncInserter.rowsBuffer) && !empty; {
		select {
		case element := <-syncInserter.queue:
			syncInserter.notifiesBuffer = append(syncInserter.notifiesBuffer, element.notifyCh)
			syncInserter.rowsBuffer = append(syncInserter.rowsBuffer, element.row)
		default:
			empty = true
		}
	}

	if len(syncInserter.rowsBuffer) == 0 {
		return
	}

	err := syncInserter.batchInsertFunc(syncInserter.rowsBuffer)

	for _, notifyCh := range syncInserter.notifiesBuffer {
		notifyCh <- err
	}

	syncInserter.rowsBuffer = syncInserter.rowsBuffer[:0]
	syncInserter.notifiesBuffer = syncInserter.notifiesBuffer[:0]
}
