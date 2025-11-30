# syncinserter
syncinserter is a go package which provides abstraction to merge concurrency inserts into batch insert.
====
[![Go Reference](https://pkg.go.dev/badge/image)](https://pkg.go.dev/github.com/alexdyukov/syncinserter)
[![Go Report](https://goreportcard.com/badge/github.com/alexdyukov/syncinserter)](https://goreportcard.com/report/github.com/alexdyukov/syncinserter)
[![Go Coverage](https://github.com/alexdyukov/syncinserter/wiki/coverage.svg)](https://raw.githack.com/wiki/alexdyukov/syncinserter/coverage.html)

## benchmarks
```
$ go clean -testcache && go test ./... && go test -timeout 5m -bench=. -benchtime=1000000x -benchmem ./...
ok      github.com/alexdyukov/syncinserter      4.476s
goos: linux
goarch: amd64
pkg: github.com/alexdyukov/syncinserter
cpu: AMD Ryzen 7 8845H w/ Radeon 780M Graphics
BenchmarkPostgres-16             1000000             13897 ns/op             973 B/op         23 allocs/op
BenchmarkClickhouse-16           1000000             24904 ns/op             307 B/op          5 allocs/op
BenchmarkOverhead-16             1000000              5258 ns/op              73 B/op          0 allocs/op
PASS
ok      github.com/alexdyukov/syncinserter      69.565s
```

## License
MIT licensed. See the included LICENSE file for details.
