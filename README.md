# syncinserter
syncinserter is a go package which provides abstraction to merge concurrency inserts into batch insert.
====
[![GoDoc](https://godoc.org/github.com/alexdyukov/syncinserter?status.svg)](https://godoc.org/github.com/alexdyukov/syncinserter)
[![Tests](https://github.com/alexdyukov/syncinserter/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/alexdyukov/ratelimiter/actions/workflows/test.yml?query=branch%3Amaster)
[![Coverage](https://github.com/alexdyukov/syncinserter/wiki/coverage.svg)](https://raw.githack.com/wiki/alexdyukov/syncinserter/coverage.html)

## benchmarks
```
$ go clean -testcache && go test ./... && go test -timeout 5m -bench=. -benchtime=100000x -benchmem ./...
ok      github.com/alexdyukov/syncinserter      3.682s
goos: linux
goarch: amd64
pkg: github.com/alexdyukov/syncinserter
cpu: AMD Ryzen 7 8845H w/ Radeon 780M Graphics
BenchmarkPostgres-16              100000             19967 ns/op            1817 B/op         30 allocs/op
BenchmarkClickhouse-16            100000             61006 ns/op             627 B/op         10 allocs/op
BenchmarkOverhead-16              100000              7996 ns/op             467 B/op          7 allocs/op
PASS
ok      github.com/alexdyukov/syncinserter      35.876s
```

## License

MIT licensed. See the included LICENSE file for details.