package syncinserter_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/alexdyukov/syncinserter/v2"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"

	cassandraTestContainer "github.com/testcontainers/testcontainers-go/modules/cassandra"
	clickhouseTestContainer "github.com/testcontainers/testcontainers-go/modules/clickhouse"
	postgresTestContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/goleak"
)

type row struct {
	createdAt time.Time
	usr       uuid.UUID
	diff      float64
}

var uuids []uuid.UUID

func TestMain(m *testing.M) {
	const uuidCount = 5000

	uuids = make([]uuid.UUID, 0, uuidCount)

	for range uuidCount {
		uuids = append(uuids, uuid.New())
	}

	os.Exit(m.Run())
}

func TestInvalidParameters(t *testing.T) {
	defer goleak.VerifyNone(t)

	insertFunc := func(_ [][]any) error { return nil }

	_, err := syncinserter.New(t.Context(), insertFunc, 0, time.Duration(1))
	if !errors.Is(err, syncinserter.ErrInvalidMaxBatchSize) {
		t.Fatalf("ErrInvalidMaxBatchSize expected, but got %s", err.Error())
	}

	_, err = syncinserter.New(t.Context(), insertFunc, 1, 0)
	if !errors.Is(err, syncinserter.ErrInvalidPeriod) {
		t.Fatalf("ErrInvalidPeriod expected, but got %s", err.Error())
	}
}

func TestCanceledContextInsert(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	insertFunc := func(_ [][]any) error { return nil }

	inserter, err := syncinserter.New(ctx, insertFunc, 1, time.Duration(1))
	if err != nil {
		t.Fatal(err.Error())
	}

	testContext, cancelTestContext := context.WithCancel(t.Context())
	cancelTestContext()

	err = inserter.Insert(testContext, []any{})
	if !errors.Is(err, testContext.Err()) {
		t.Fatalf("context error expected, but got %s", err.Error())
	}
}

func TestSyncInserter(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	testName := strings.ToLower(t.Name())

	container, err := postgresTestContainer.Run(ctx, "postgres:latest", testcontainers.WithName(testName), testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")))
	if err != nil {
		t.Fatal(err.Error())
	}

	defer container.Terminate(ctx)

	connString, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		t.Fatal(err.Error())
	}

	conn, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err.Error())
	}

	defer conn.Close()

	err = conn.Ping(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS `+testName+` (created_at TIMESTAMP, usr UUID, diff double precision);`)
	if err != nil {
		t.Fatal(err.Error())
	}

	insertFunc := func(rows [][]any) error {
		_, err := conn.CopyFrom(context.Background(), pgx.Identifier{testName}, []string{"created_at", "usr", "diff"}, pgx.CopyFromRows(rows))

		return err
	}

	const batchSize = 1000

	inserter, err := syncinserter.New(ctx, insertFunc, batchSize, time.Duration(1))
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

	err = conn.QueryRow(ctx, `SELECT created_at, diff FROM `+testName+` WHERE usr = $1;`, insertedUserID).Scan(&selectedCreatedAt, &selectedValue)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !insertedCreatedAt.Equal(selectedCreatedAt) || math.Abs(insertedValue-selectedValue) > 0.0001 {
		t.Fatalf("invalid data in %s: want %v, %v but got %v, %v", testName, insertedCreatedAt, insertedValue, selectedCreatedAt, selectedValue)
	}
}

func BenchmarkOverhead(b *testing.B) {
	// https://github.com/testcontainers/testcontainers-go/issues/2878
	// defer goleak.VerifyNone(b)
	b.StopTimer()

	ctx, cancel := context.WithCancel(b.Context())
	defer cancel()

	insertFunc := func(rows []row) error { return nil }

	const batchSize = 1000

	inserter, err := syncinserter.New(ctx, insertFunc, batchSize, time.Duration(1))
	if err != nil {
		b.Fatal(err.Error())
	}

	b.StartTimer()

	for b.Loop() {
		err = inserter.Insert(ctx, row{})
		if err != nil {
			b.Fatal(err.Error())
		}
	}
}

func BenchmarkPostgres(b *testing.B) {
	// https://github.com/testcontainers/testcontainers-go/issues/2878
	// defer goleak.VerifyNone(b)
	b.StopTimer()

	testName := strings.ToLower(b.Name())

	container, err := postgresTestContainer.Run(
		b.Context(),
		"postgres:latest",
		testcontainers.WithName(testName),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		b.Fatal(err.Error())
	}

	defer container.Terminate(context.Background())

	connString, err := container.ConnectionString(b.Context())
	if err != nil {
		b.Fatal(err.Error())
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		b.Fatal(err.Error())
	}

	conn, err := pgxpool.NewWithConfig(b.Context(), config)
	if err != nil {
		b.Fatal(err.Error())
	}

	defer conn.Close()

	err = conn.Ping(b.Context())
	if err != nil {
		b.Fatal(err.Error())
	}

	_, err = conn.Exec(b.Context(), `CREATE TABLE IF NOT EXISTS `+testName+` (created_at TIMESTAMP, usr UUID, diff double precision);`)
	if err != nil {
		b.Fatal(err.Error())
	}

	insertFunc := func(rows [][]any) error {
		_, err := conn.CopyFrom(b.Context(), pgx.Identifier{testName}, []string{"created_at", "usr", "diff"}, pgx.CopyFromRows(rows))

		return err
	}

	batchSizes := []int{100, 1000, 10000}
	for _, batchSize := range batchSizes {
		for parallelism := batchSize / 10; parallelism <= batchSize*10 && parallelism <= batchSizes[len(batchSizes)-1]; parallelism *= 10 {
			b.Run(fmt.Sprintf("BatchSize%dParallelism%d", batchSize, parallelism), func(b *testing.B) {
				ctx, cancel := context.WithCancel(b.Context())
				defer cancel()

				inserter, err := syncinserter.New(ctx, insertFunc, batchSize, time.Duration(1))
				if err != nil {
					b.Fatal(err.Error())
				}

				readyCh := make(chan struct{})
				closeCh := make(chan struct{})

				for range parallelism - 1 {
					go func() {
						<-readyCh

						for {
							select {
							case <-closeCh:
								return
							default:
								_ = inserter.Insert(ctx, []any{time.Now().UTC(), uuids[rand.Intn(len(uuids))], rand.Float64()})
							}
						}
					}()
				}

				b.StartTimer()

				close(readyCh)
				defer close(closeCh)

				for b.Loop() {
					err = inserter.Insert(ctx, []any{time.Now().UTC(), uuids[rand.Intn(len(uuids))], rand.Float64()})
					if err != nil {
						b.Fatal(err.Error())
					}
				}

				b.StopTimer()
			})
		}
	}
}

func BenchmarkClickhouse(b *testing.B) {
	// https://github.com/testcontainers/testcontainers-go/issues/2878
	// defer goleak.VerifyNone(b)
	b.StopTimer()

	testName := strings.ToLower(b.Name())

	container, err := clickhouseTestContainer.Run(
		b.Context(),
		"clickhouse/clickhouse-server:latest",
		testcontainers.WithName(testName),
		clickhouseTestContainer.WithUsername(testName),
		clickhouseTestContainer.WithPassword(testName),
	)
	if err != nil {
		b.Fatal(err.Error())
	}

	defer container.Terminate(context.Background())

	connString, err := container.ConnectionString(b.Context())
	if err != nil {
		b.Fatal(err.Error())
	}

	config, err := clickhouse.ParseDSN(connString)
	if err != nil {
		b.Fatal(err.Error())
	}

	conn, err := clickhouse.Open(config)
	if err != nil {
		b.Fatal(err.Error())
	}

	defer conn.Close()

	err = conn.Ping(b.Context())
	if err != nil {
		b.Fatal(err.Error())
	}

	err = conn.Exec(b.Context(), `CREATE TABLE IF NOT EXISTS `+testName+` (created_at DateTime64(9, 'UTC'), usr UUID, diff Float64) ENGINE = MergeTree() ORDER BY created_at;`)
	if err != nil {
		b.Fatal(err.Error())
	}

	insertFunc := func(rows []row) error {
		batch, err := conn.PrepareBatch(b.Context(), `INSERT INTO `+testName+` (created_at, usr, diff)`)
		if err != nil {
			return err
		}
		defer batch.Close()

		for _, row := range rows {
			err = batch.Append(row.createdAt, row.usr, row.diff)
			if err != nil {
				return err
			}
		}

		return batch.Send()
	}

	// https://clickhouse.com/docs/optimize/bulk-inserts
	// We recommend inserting data in batches of at least 1,000 rows, and ideally between 10,000–100,000 rows.
	// Fewer, larger inserts reduce the number of parts written, minimize merge load, and lower overall system resource usage.
	batchSizes := []int{1000, 10000, 100000}
	for _, batchSize := range batchSizes {
		for parallelism := batchSize / 10; parallelism <= batchSize*10 && parallelism <= batchSizes[len(batchSizes)-1]; parallelism *= 10 {
			b.Run(fmt.Sprintf("BatchSize%dParallelism%d", batchSize, parallelism), func(b *testing.B) {
				ctx, cancel := context.WithCancel(b.Context())
				defer cancel()

				inserter, err := syncinserter.New(ctx, insertFunc, batchSize, time.Duration(1))
				if err != nil {
					b.Fatal(err.Error())
				}

				readyCh := make(chan struct{})
				closeCh := make(chan struct{})

				for range parallelism - 1 {
					go func() {
						<-readyCh

						for {
							select {
							case <-closeCh:
								return
							default:
								r := row{
									createdAt: time.Now().UTC(),
									usr:       uuids[rand.Intn(len(uuids))],
									diff:      rand.Float64(),
								}

								_ = inserter.Insert(ctx, r)
							}
						}
					}()
				}

				b.StartTimer()

				close(readyCh)
				defer close(closeCh)

				for b.Loop() {
					r := row{
						createdAt: time.Now().UTC(),
						usr:       uuids[rand.Intn(len(uuids))],
						diff:      rand.Float64(),
					}

					err = inserter.Insert(ctx, r)
					if err != nil {
						b.Fatal(err.Error())
					}
				}

				b.StopTimer()
			})
		}
	}
}

func BenchmarkCassandra(b *testing.B) {
	// https://github.com/testcontainers/testcontainers-go/issues/2878
	// defer goleak.VerifyNone(b)
	b.StopTimer()

	ctx, cancel := context.WithCancel(b.Context())
	defer cancel()

	testName := strings.ToLower(b.Name())

	container, err := cassandraTestContainer.Run(
		ctx,
		"cassandra:latest",
		testcontainers.WithName(testName),
		cassandraTestContainer.WithInitScripts(filepath.Join("testdata", "cassandra_init.cql")),
	)
	if err != nil {
		b.Fatal(err.Error())
	}

	defer container.Terminate(ctx)

	connHost, err := container.ConnectionHost(ctx)
	if err != nil {
		b.Fatal(err.Error())
	}

	connConfig := gocql.NewCluster(connHost)
	connConfig.Keyspace = "testkeyspace"

	session, err := connConfig.CreateSession()
	if err != nil {
		b.Fatal(err.Error())
	}

	defer session.Close()

	err = session.Query(`CREATE TABLE IF NOT EXISTS ` + testName + ` (created_at TIMESTAMP, usr UUID, diff DOUBLE, PRIMARY KEY (created_at, usr));`).Exec()
	if err != nil {
		b.Fatal(err.Error())
	}

	insertFunc := func(rows []row) error {
		batch := session.NewBatch(gocql.LoggedBatch)

		insertStr := `INSERT INTO ` + testName + ` (created_at, usr, diff) VALUES (?, ?, ?)`

		for _, row := range rows {
			batch.Query(insertStr, row.createdAt, row.usr.String(), row.diff)
		}

		return session.ExecuteBatch(batch)
	}

	batchSizes := []int{100, 250, 500}
	for _, batchSize := range batchSizes {
		for parallelism := batchSize / 10; parallelism <= batchSize*10 && parallelism <= batchSizes[len(batchSizes)-1]; parallelism *= 10 {
			b.Run(fmt.Sprintf("BatchSize%dParallelism%d", batchSize, parallelism), func(b *testing.B) {
				ctx, cancel := context.WithCancel(b.Context())
				defer cancel()

				inserter, err := syncinserter.New(ctx, insertFunc, batchSize, time.Duration(1))
				if err != nil {
					b.Fatal(err.Error())
				}

				readyCh := make(chan struct{})
				closeCh := make(chan struct{})

				for range parallelism - 1 {
					go func() {
						<-readyCh

						for {
							select {
							case <-closeCh:
								return
							default:
								r := row{
									createdAt: time.Now().UTC(),
									usr:       uuids[rand.Intn(len(uuids))],
									diff:      rand.Float64(),
								}

								_ = inserter.Insert(ctx, r)
							}
						}
					}()
				}

				b.StartTimer()

				close(readyCh)
				defer close(closeCh)

				for b.Loop() {
					r := row{
						createdAt: time.Now().UTC(),
						usr:       uuids[rand.Intn(len(uuids))],
						diff:      rand.Float64(),
					}

					err = inserter.Insert(ctx, r)
					if err != nil {
						b.Fatal(err.Error())
					}
				}

				b.StopTimer()
			})
		}
	}
}
