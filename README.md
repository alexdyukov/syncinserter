# syncinserter
syncinserter is a go package which provides abstraction to merge concurrency inserts into batch insert.
====
[![Go Reference](https://pkg.go.dev/badge/image)](https://pkg.go.dev/github.com/alexdyukov/syncinserter/v2)
[![Go Coverage](https://github.com/alexdyukov/syncinserter/wiki/coverage.svg)](https://raw.githack.com/wiki/alexdyukov/syncinserter/coverage.html)

## benchmarks
```
$ go clean -testcache && go test ./... && go test -timeout 5m -bench=. -benchtime=100x -benchmem ./...
ok      github.com/alexdyukov/syncinserter/v2   4.191s
goos: linux
goarch: amd64
pkg: github.com/alexdyukov/syncinserter/v2
cpu: AMD Ryzen 7 8845H w/ Radeon 780M Graphics
BenchmarkOverhead-16                                         100             47707 ns/op              28 B/op          0 allocs/op
BenchmarkPostgres/BatchSize100Parallelism10-16               100           1027562 ns/op           13868 B/op        311 allocs/op
BenchmarkPostgres/BatchSize100Parallelism100-16              100           1584716 ns/op          104255 B/op       2502 allocs/op
BenchmarkPostgres/BatchSize100Parallelism1000-16             100          10506109 ns/op          996959 B/op      24461 allocs/op
BenchmarkPostgres/BatchSize1000Parallelism100-16             100           1481120 ns/op           99928 B/op       2450 allocs/op
BenchmarkPostgres/BatchSize1000Parallelism1000-16            100          10710141 ns/op          926072 B/op      24193 allocs/op
BenchmarkPostgres/BatchSize1000Parallelism10000-16           100         127072101 ns/op         8953991 B/op     239400 allocs/op
BenchmarkPostgres/BatchSize10000Parallelism1000-16           100           9181152 ns/op          896388 B/op      23783 allocs/op
BenchmarkPostgres/BatchSize10000Parallelism10000-16          100         114327203 ns/op         7561784 B/op     241359 allocs/op
BenchmarkClickhouse/BatchSize1000Parallelism100-16           100         113167542 ns/op           47229 B/op        795 allocs/op
BenchmarkClickhouse/BatchSize1000Parallelism1000-16          100          87545424 ns/op          235013 B/op       4305 allocs/op
BenchmarkClickhouse/BatchSize1000Parallelism10000-16         100         704885409 ns/op         2117934 B/op      42382 allocs/op
BenchmarkClickhouse/BatchSize10000Parallelism1000-16         100          87749603 ns/op          221979 B/op       4282 allocs/op
BenchmarkClickhouse/BatchSize10000Parallelism10000-16        100         178229064 ns/op         2610375 B/op      42045 allocs/op
BenchmarkClickhouse/BatchSize10000Parallelism100000-16       100        1578430467 ns/op        23933588 B/op     403621 allocs/op
BenchmarkClickhouse/BatchSize100000Parallelism10000-16       100         174787536 ns/op         2383434 B/op      40111 allocs/op
BenchmarkClickhouse/BatchSize100000Parallelism100000-16      100        1016842729 ns/op        29257759 B/op     421067 allocs/op
BenchmarkCassandra/BatchSize100Parallelism10-16              100           7790287 ns/op           12679 B/op        187 allocs/op
BenchmarkCassandra/BatchSize100Parallelism100-16             100           7866495 ns/op           99767 B/op       1153 allocs/op
BenchmarkCassandra/BatchSize250Parallelism25-16              100           4585951 ns/op           27479 B/op        358 allocs/op
BenchmarkCassandra/BatchSize250Parallelism250-16             100           7266076 ns/op          246850 B/op       2782 allocs/op
BenchmarkCassandra/BatchSize500Parallelism50-16              100           4565652 ns/op           53882 B/op        641 allocs/op
BenchmarkCassandra/BatchSize500Parallelism500-16             100          10021854 ns/op          522635 B/op       5509 allocs/op
PASS
ok      github.com/alexdyukov/syncinserter/v2   470.954s
```

## License
MIT licensed. See the included LICENSE file for details.
