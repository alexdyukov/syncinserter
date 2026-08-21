package syncinserter_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/alexdyukov/syncinserter/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	cockroachDBTestContainer "github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func prepareCockroachDB(t test) (*pgxpool.Pool, func(), error) {
	t.Helper()

	cleanup := func() {}

	container, err := cockroachDBTestContainer.Run(
		t.Context(),
		"cockroachdb/cockroach",
		cockroachDBTestContainer.WithInsecure(),
	)
	if err != nil {
		fmt.Println(1)
		return nil, cleanup, err
	}

	cleanup = func() {
		_ = container.Terminate(context.Background())
	}

	connConfig, err := container.ConnectionConfig(t.Context())
	if err != nil {
		fmt.Println(2)
		return nil, cleanup, err
	}

	config, err := pgxpool.ParseConfig(connConfig.ConnString())
	if err != nil {
		fmt.Println(3)
		return nil, cleanup, err
	}

	config.MaxConns = int32(dbMaxConnection)
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	conn, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		return nil, cleanup, err
	}

	cleanup = func() {
		conn.Close()
		_ = container.Terminate(context.Background())
	}

	_, err = conn.Exec(t.Context(), `CREATE TABLE test (created_at TIMESTAMP, usr UUID, diff double precision);`)

	return conn, cleanup, err
}

func BenchmarkCockroachDBDirectInsert(b *testing.B) {
	for _, parallelism := range parallelismPlan {
		b.Run(fmt.Sprintf("Concurrency%d", parallelism*goMaxProcs), func(sb *testing.B) {
			sb.StopTimer()

			conn, cleanup, err := prepareCockroachDB(sb)
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
					_, err = conn.Exec(sb.Context(), sqlQuery, time.Now().UTC(), uuids[rand.Intn(len(uuids))], rand.Float64())
					if err != nil {
						sb.Fatal(err.Error())
					}
				}
			})
		})
	}
}

func BenchmarkCockroachDBWrappedMultiline(b *testing.B) {
	for _, parallelism := range parallelismPlan {
		for _, batchSize := range batchSizePlan {
			b.Run(fmt.Sprintf("Concurrency%dBatchSize%d", parallelism*goMaxProcs, batchSize), func(sb *testing.B) {
				sb.StopTimer()

				conn, cleanup, err := prepareCockroachDB(sb)
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

					_, err := conn.Exec(sb.Context(), insertStr, params...)

					return err
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
