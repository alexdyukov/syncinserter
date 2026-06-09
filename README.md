# syncinserter
syncinserter is a go package which provides abstraction to merge concurrency inserts into batch insert.
====
[![Go Reference](https://pkg.go.dev/badge/image)](https://pkg.go.dev/github.com/alexdyukov/syncinserter)
[![Go Report](https://goreportcard.com/badge/github.com/alexdyukov/syncinserter)](https://goreportcard.com/report/github.com/alexdyukov/syncinserter)
[![Go Coverage](https://github.com/alexdyukov/syncinserter/wiki/coverage.svg)](https://raw.githack.com/wiki/alexdyukov/syncinserter/coverage.html)

## benchmarks
```
$ go clean -testcache && go test ./... && go test -timeout 5m -bench=. -benchtime=1000000x -benchmem ./...
ok      github.com/alexdyukov/syncinserter      4.538s
goos: linux
goarch: amd64
pkg: github.com/alexdyukov/syncinserter
cpu: AMD Ryzen 7 8845H w/ Radeon 780M Graphics
BenchmarkOverhead-16             1000000              6144 ns/op             156 B/op          1 allocs/op
BenchmarkPostgres-16             1000000             14973 ns/op             898 B/op         23 allocs/op
BenchmarkClickhouse-16           1000000             19956 ns/op             311 B/op          4 allocs/op
BenchmarkCassandra-16            1000000             33201 ns/op            1081 B/op         11 allocs/op
PASS
ok      github.com/alexdyukov/syncinserter      128.664s
```

## License
MIT licensed. See the included LICENSE file for details.
