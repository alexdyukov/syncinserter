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
	queueElement struct {
		notifyCh chan error
		row      []any
	}

	// SyncInserter is bulk inserter with sync Insert API. It merges all Insert calls during period and merges them in one batch call.
	SyncInserter struct {
		notifiesBuffer []chan error
		rowsBuffer     [][]any
		batchInsert    func(rows [][]any) error
		queue          chan queueElement
	}
)

// New initializes SyncInserter.
func New(ctx context.Context, batchInsert func(rows [][]any) error, maxBatchSize int, period time.Duration) (*SyncInserter, error) {
	if maxBatchSize <= 0 {
		return nil, ErrInvalidMaxBatchSize
	}

	if period <= 0 {
		return nil, ErrInvalidPeriod
	}

	syncInserter := &SyncInserter{
		batchInsert:    batchInsert,
		queue:          make(chan queueElement),
		notifiesBuffer: make([]chan error, 0, maxBatchSize),
		rowsBuffer:     make([][]any, 0, maxBatchSize),
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
func (syncInserter *SyncInserter) Insert(ctx context.Context, row []any) error {
	notifyCh := notifyChannelPool.Get().(chan error)

	select {
	case <-ctx.Done():
		notifyChannelPool.Put(notifyCh)

		return ctx.Err()
	case syncInserter.queue <- queueElement{row: row, notifyCh: notifyCh}:
		return <-notifyCh
	}
}

func (syncInserter *SyncInserter) send() {
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

	err := syncInserter.batchInsert(syncInserter.rowsBuffer)

	for _, notifyCh := range syncInserter.notifiesBuffer {
		notifyCh <- err

		notifyChannelPool.Put(notifyCh)
	}

	syncInserter.notifiesBuffer = syncInserter.notifiesBuffer[:0]
	syncInserter.rowsBuffer = syncInserter.rowsBuffer[:0]
}
