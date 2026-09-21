package syncinserter_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/alexdyukov/syncinserter/v2"
	clickhouseTestContainer "github.com/testcontainers/testcontainers-go/modules/clickhouse"
)

func prepareClickhouse(t test) (driver.Conn, func(), error) {
	t.Helper()

	cleanup := func() {}

	container, err := clickhouseTestContainer.Run(
		t.Context(),
		"clickhouse/clickhouse-server",
		clickhouseTestContainer.WithUsername("test"),
		clickhouseTestContainer.WithPassword("test"),
	)
	if err != nil {
		return nil, cleanup, err
	}

	cleanup = func() {
		_ = container.Terminate(context.Background())
	}

	connString, err := container.ConnectionString(t.Context())
	if err != nil {
		return nil, cleanup, err
	}

	config, err := clickhouse.ParseDSN(connString)
	if err != nil {
		return nil, cleanup, err
	}

	config.MaxOpenConns = dbMaxConnection

	conn, err := clickhouse.Open(config)
	if err != nil {
		return nil, cleanup, err
	}

	err = conn.Ping(t.Context())
	if err != nil {
		return nil, cleanup, err
	}

	cleanup = func() {
		_ = conn.Close()
		_ = container.Terminate(context.Background())
	}

	err = conn.Exec(t.Context(), `CREATE TABLE test (created_at DateTime64(9, 'UTC'), usr UUID, diff Float64) ENGINE = MergeTree() ORDER BY created_at;`)

	return conn, cleanup, err
}

func BenchmarkClickhouseDirectInsert(b *testing.B) {
	for _, parallelism := range parallelismPlan {
		b.Run(fmt.Sprintf("Concurrency%d", parallelism*goMaxProcs), func(sb *testing.B) {
			sb.StopTimer()

			conn, cleanup, err := prepareClickhouse(sb)
			defer cleanup()

			if err != nil {
				sb.Fatal(err.Error())
			}

			sqlQuery := `INSERT INTO test (created_at, usr, diff) VALUES ($1, $2, $3);`

			sb.SetParallelism(parallelism)

			sb.StartTimer()
			defer sb.StopTimer()

			sb.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					err := conn.Exec(sb.Context(), sqlQuery, time.Now().UTC(), uuids[rand.Intn(len(uuids))], rand.Float64())
					if err != nil {
						sb.Fatal(err.Error())
					}
				}
			})
		})
	}
}

func BenchmarkClickhouseWrappedBatch(b *testing.B) {
	for _, parallelism := range parallelismPlan {
		for _, batchSize := range batchSizePlan {
			b.Run(fmt.Sprintf("Concurrency%dBatchSize%d", parallelism*goMaxProcs, batchSize), func(sb *testing.B) {
				sb.StopTimer()

				conn, cleanup, err := prepareClickhouse(sb)
				defer cleanup()

				if err != nil {
					sb.Fatal(err.Error())
				}

				insertFunc := func(rows []row) error {
					batch, err := conn.PrepareBatch(b.Context(), `INSERT INTO test (created_at, usr, diff)`)
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

				inserter, err := syncinserter.New(sb.Context(), insertFunc, batchSize, time.Duration(1))
				if err != nil {
					sb.Fatal(err.Error())
				}

				sb.SetParallelism(parallelism)

				sb.StartTimer()
				defer sb.StopTimer()

				sb.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						err := inserter.Insert(sb.Context(), row{time.Now().UTC(), uuids[rand.Intn(len(uuids))], rand.Float64()})
						if err != nil {
							sb.Fatal(err.Error())
						}
					}
				})
			})
		}
	}
}

func BenchmarkClickhouseWrappedMultiline(b *testing.B) {
	for _, parallelism := range parallelismPlan {
		for _, batchSize := range batchSizePlan {
			b.Run(fmt.Sprintf("Concurrency%dBatchSize%d", parallelism*goMaxProcs, batchSize), func(sb *testing.B) {
				sb.StopTimer()

				conn, cleanup, err := prepareClickhouse(sb)
				defer cleanup()

				if err != nil {
					sb.Fatal(err.Error())
				}

				insertFunc := func(rows []row) error {
					insertStr := insertString(len(rows))

					params := make([]any, 0, len(rows)*3)
					for _, row := range rows {
						params = append(params, row.createdAt, row.usr.String(), row.diff)
					}

					return conn.Exec(sb.Context(), insertStr, params...)
				}

				inserter, err := syncinserter.New(sb.Context(), insertFunc, batchSize, time.Duration(1))
				if err != nil {
					sb.Fatal(err.Error())
				}

				sb.SetParallelism(parallelism)

				sb.StartTimer()
				defer sb.StopTimer()

				sb.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						err := inserter.Insert(sb.Context(), row{time.Now().UTC(), uuids[rand.Intn(len(uuids))], rand.Float64()})
						if err != nil {
							sb.Fatal(err.Error())
						}
					}
				})
			})
		}
	}
}
