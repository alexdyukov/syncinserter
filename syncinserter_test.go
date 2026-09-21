package syncinserter_test

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/alexdyukov/syncinserter/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	testcontainersLog "github.com/testcontainers/testcontainers-go/log"
	"go.uber.org/goleak"
)

type test interface {
	Helper()
	Context() context.Context
}

type row struct {
	createdAt time.Time
	usr       uuid.UUID
	diff      float64
}

var (
	uuids           []uuid.UUID
	parallelismPlan = []int{5, 15, 25}
	batchSizePlan   = []int{50, 150, 500}
	goMaxProcs      = runtime.GOMAXPROCS(0)
	dbMaxConnection = runtime.GOMAXPROCS(0) * 2
)

func TestMain(m *testing.M) {
	const uuidCount = 5000

	uuids = make([]uuid.UUID, 0, uuidCount)

	for range uuidCount {
		uuids = append(uuids, uuid.New())
	}

	testcontainersLog.SetDefault(testcontainersLog.NewNoopLogger())

	os.Exit(m.Run())
}

func TestInvalidParameters(t *testing.T) {
	defer goleak.VerifyNone(t)

	insertFunc := func(_ []row) error { return nil }

	const (
		validBatch    = 1
		invalidBatch  = 0
		validPeriod   = time.Duration(1)
		invalidPeriod = time.Duration(0)
	)

	_, err := syncinserter.New(t.Context(), insertFunc, invalidBatch, validPeriod)
	if !errors.Is(err, syncinserter.ErrInvalidMaxBatchSize) {
		t.Fatalf("ErrInvalidMaxBatchSize expected, but got %s", err.Error())
	}

	_, err = syncinserter.New(t.Context(), insertFunc, validBatch, invalidPeriod)
	if !errors.Is(err, syncinserter.ErrInvalidPeriod) {
		t.Fatalf("ErrInvalidPeriod expected, but got %s", err.Error())
	}
}

func TestCanceledContextInsert(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	insertFunc := func(_ []row) error { return nil }

	const (
		batchSize = 1
		period    = time.Duration(1)
	)

	inserter, err := syncinserter.New(ctx, insertFunc, batchSize, period)
	if err != nil {
		t.Fatal(err.Error())
	}

	testContext, cancelTestContext := context.WithCancel(t.Context())
	cancelTestContext()

	err = inserter.Insert(testContext, row{})
	if !errors.Is(err, testContext.Err()) {
		t.Fatalf("context error expected, but got %s", err.Error())
	}
}

func TestSyncInserter(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	conn, cleanup, err := preparePostgres(t)
	defer cleanup()

	if err != nil {
		t.Fatal(err.Error())
	}

	insertFunc := func(rows [][]any) error {
		_, err := conn.CopyFrom(context.Background(), pgx.Identifier{"test"}, []string{"created_at", "usr", "diff"}, pgx.CopyFromRows(rows))

		return err
	}

	const (
		batchSize = 1000
		period    = time.Duration(1)
	)

	inserter, err := syncinserter.New(ctx, insertFunc, batchSize, period)
	if err != nil {
		t.Fatal(err.Error())
	}

	// round to microsecond, cause datetime postgres resolution
	// https://www.postgresql.org/docs/current/datatype-datetime.html
	insertedCreatedAt := time.Now().UTC().Round(time.Microsecond)
	insertedUserID := uuid.New().String()
	insertedValue := rand.Float64()

	err = inserter.Insert(ctx, []any{insertedCreatedAt, insertedUserID, insertedValue})
	if err != nil {
		t.Fatal(err.Error())
	}

	var (
		selectedCreatedAt time.Time
		selectedValue     float64
	)

	err = conn.QueryRow(ctx, `SELECT created_at, diff FROM test WHERE usr = $1;`, insertedUserID).Scan(&selectedCreatedAt, &selectedValue)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !insertedCreatedAt.Equal(selectedCreatedAt) || math.Abs(insertedValue-selectedValue) > 0.0001 {
		t.Fatalf("invalid data in table: want %v, %v but got %v, %v", insertedCreatedAt, insertedValue, selectedCreatedAt, selectedValue)
	}
}
