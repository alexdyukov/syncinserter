# syncinserter
syncinserter is a go package which provides abstraction to merge concurrency inserts into batch insert.
====
[![Go Reference](https://pkg.go.dev/badge/image)](https://pkg.go.dev/github.com/alexdyukov/syncinserter)
[![Go Report](https://goreportcard.com/badge/github.com/alexdyukov/syncinserter)](https://goreportcard.com/report/github.com/alexdyukov/syncinserter)
[![Go Coverage](https://github.com/alexdyukov/syncinserter/wiki/coverage.svg)](https://raw.githack.com/wiki/alexdyukov/syncinserter/coverage.html)

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
