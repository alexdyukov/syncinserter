package syncinserter_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/alexdyukov/syncinserter/v2"
	"github.com/gocql/gocql"

	"github.com/testcontainers/testcontainers-go"
	scylladbTestContainer "github.com/testcontainers/testcontainers-go/modules/scylladb"
	"github.com/testcontainers/testcontainers-go/wait"
)

func initScyllaDBKeyspace(connConfig *gocql.ClusterConfig, keyspace string) error {
	session, err := connConfig.CreateSession()
	if err != nil {
		return err
	}
	defer session.Close()

	q := session.Query(`CREATE KEYSPACE ` + keyspace + ` WITH REPLICATION = {'class':'NetworkTopologyStrategy','replication_factor':1};`)
	defer q.Release()

	return q.Exec()
}

func prepareScyllaDB(t test) (*gocql.Session, func(), error) {
	t.Helper()

	cleanup := func() {}

	container, err := scylladbTestContainer.Run(
		t.Context(),
		"scylladb/scylla",
		testcontainers.WithWaitStrategy(wait.ForListeningPort("9042/tcp")),
	)
	if err != nil {
		return nil, cleanup, err
	}

	cleanup = func() {
		_ = container.Terminate(context.Background())
	}

	connHost, err := container.NonShardAwareConnectionHost(t.Context())
	if err != nil {
		return nil, cleanup, err
	}

	connConfig := gocql.NewCluster(connHost)

	const keyspace = "testkeyspace"

	err = initScyllaDBKeyspace(connConfig, keyspace)
	if err != nil {
		return nil, cleanup, err
	}

	connConfig.NumConns = dbMaxConnection
	connConfig.Keyspace = keyspace

	conn, err := connConfig.CreateSession()
	if err != nil {
		return conn, cleanup, err
	}

	cleanup = func() {
		conn.Close()
		_ = container.Terminate(context.Background())
	}

	q := conn.Query(`CREATE TABLE test (created_at TIMESTAMP, usr UUID, diff DOUBLE, PRIMARY KEY (created_at, usr));`)
	defer q.Release()

	return conn, cleanup, q.Exec()
}

func BenchmarkScyllaDBDirectInsert(b *testing.B) {
	for _, parallelism := range parallelismPlan {
		b.Run(fmt.Sprintf("Concurrency%d", parallelism*goMaxProcs), func(sb *testing.B) {
			sb.StopTimer()

			conn, cleanup, err := prepareScyllaDB(sb)
			defer cleanup()

			if err != nil {
				sb.Fatal(err.Error())
			}

			sqlQuery := `INSERT INTO test (created_at, usr, diff) VALUES (?, ?, ?);`

			sb.SetParallelism(parallelism)
			sb.StartTimer()
			defer sb.StopTimer()

			sb.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					q := conn.Query(sqlQuery, time.Now().UTC(), uuids[rand.Intn(len(uuids))].String(), rand.Float64())
					err = q.Exec()
					q.Release()
					if err != nil {
						sb.Fatal(err.Error())
					}
				}
			})
		})
	}
}

func BenchmarkScyllaDBWrappedBatch(b *testing.B) {
	for _, parallelism := range parallelismPlan {
		for _, batchSize := range batchSizePlan {
			b.Run(fmt.Sprintf("Concurrency%dBatchSize%d", parallelism*goMaxProcs, batchSize), func(sb *testing.B) {
				sb.StopTimer()

				conn, cleanup, err := prepareScyllaDB(sb)
				defer cleanup()

				if err != nil {
					sb.Fatal(err.Error())
				}

				insertFunc := func(rows []row) error {
					batch := conn.NewBatch(gocql.UnloggedBatch)

					insertStr := `INSERT INTO test (created_at, usr, diff) VALUES (?, ?, ?);`

					for _, row := range rows {
						batch.Query(insertStr, row.createdAt, row.usr.String(), row.diff)
					}

					return conn.ExecuteBatch(batch)
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
