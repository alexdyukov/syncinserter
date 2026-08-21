# syncinserter
syncinserter is a go package which provides abstraction to merge concurrency inserts into batch insert.
====
[![Go Reference](https://pkg.go.dev/badge/image)](https://pkg.go.dev/github.com/alexdyukov/syncinserter/v2)
[![Go Coverage](https://github.com/alexdyukov/syncinserter/wiki/coverage.svg)](https://raw.githack.com/wiki/alexdyukov/syncinserter/coverage.html)

## benchmarks
```
$ go test -bench=. -benchtime=10000x -benchmem ./...
goos: linux
goarch: amd64
pkg: github.com/alexdyukov/syncinserter/v2
cpu: AMD Ryzen 7 8845H w/ Radeon 780M Graphics
BenchmarkCassandraDirectInsert/Concurrency80-16                    10000             97208 ns/op            2666 B/op         40 allocs/op
BenchmarkCassandraDirectInsert/Concurrency240-16                   10000             85853 ns/op            2694 B/op         39 allocs/op
BenchmarkCassandraDirectInsert/Concurrency400-16                   10000             76604 ns/op            2693 B/op         39 allocs/op
BenchmarkCassandraWrappedBatch/Concurrency80BatchSize50-16                 10000            111557 ns/op            1026 B/op         12 allocs/op
BenchmarkCassandraWrappedBatch/Concurrency80BatchSize150-16                10000            108599 ns/op            1038 B/op         12 allocs/op
BenchmarkCassandraWrappedBatch/Concurrency80BatchSize500-16                10000            109304 ns/op            1036 B/op         12 allocs/op
BenchmarkCassandraWrappedBatch/Concurrency240BatchSize50-16                10000             98060 ns/op            1040 B/op         12 allocs/op
BenchmarkCassandraWrappedBatch/Concurrency240BatchSize150-16               10000             81527 ns/op            1105 B/op         11 allocs/op
BenchmarkCassandraWrappedBatch/Concurrency240BatchSize500-16               10000             55913 ns/op            1031 B/op         11 allocs/op
BenchmarkCassandraWrappedBatch/Concurrency400BatchSize50-16                10000             95574 ns/op            1049 B/op         12 allocs/op
BenchmarkCassandraWrappedBatch/Concurrency400BatchSize150-16               10000             59186 ns/op            1147 B/op         11 allocs/op
BenchmarkCassandraWrappedBatch/Concurrency400BatchSize500-16               10000             49559 ns/op            1049 B/op         11 allocs/op
BenchmarkClickhouseDirectInsert/Concurrency80-16                           10000           5620070 ns/op           12592 B/op        163 allocs/op
BenchmarkClickhouseDirectInsert/Concurrency240-16                          10000           5606060 ns/op           12640 B/op        162 allocs/op
BenchmarkClickhouseDirectInsert/Concurrency400-16                          10000           5603919 ns/op           12562 B/op        162 allocs/op
BenchmarkClickhouseWrappedBatch/Concurrency80BatchSize50-16                10000           1538768 ns/op             584 B/op          9 allocs/op
BenchmarkClickhouseWrappedBatch/Concurrency80BatchSize150-16               10000           1520240 ns/op             607 B/op          9 allocs/op
BenchmarkClickhouseWrappedBatch/Concurrency80BatchSize500-16               10000           1548450 ns/op             544 B/op          9 allocs/op
BenchmarkClickhouseWrappedBatch/Concurrency240BatchSize50-16               10000           1260871 ns/op             536 B/op          8 allocs/op
BenchmarkClickhouseWrappedBatch/Concurrency240BatchSize150-16              10000            479107 ns/op             452 B/op          5 allocs/op
BenchmarkClickhouseWrappedBatch/Concurrency240BatchSize500-16              10000            395727 ns/op             375 B/op          5 allocs/op
BenchmarkClickhouseWrappedBatch/Concurrency400BatchSize50-16               10000           1254537 ns/op             499 B/op          8 allocs/op
BenchmarkClickhouseWrappedBatch/Concurrency400BatchSize150-16              10000            427942 ns/op             441 B/op          5 allocs/op
BenchmarkClickhouseWrappedBatch/Concurrency400BatchSize500-16              10000            200867 ns/op             327 B/op          4 allocs/op
BenchmarkClickhouseWrappedMultiline/Concurrency80BatchSize50-16            10000           1512972 ns/op            1737 B/op         16 allocs/op
BenchmarkClickhouseWrappedMultiline/Concurrency80BatchSize150-16           10000           1498586 ns/op            1724 B/op         16 allocs/op
BenchmarkClickhouseWrappedMultiline/Concurrency80BatchSize500-16           10000           1504283 ns/op            1664 B/op         16 allocs/op
BenchmarkClickhouseWrappedMultiline/Concurrency240BatchSize50-16           10000           1260856 ns/op            1797 B/op         16 allocs/op
BenchmarkClickhouseWrappedMultiline/Concurrency240BatchSize150-16          10000            457628 ns/op            1830 B/op         17 allocs/op
BenchmarkClickhouseWrappedMultiline/Concurrency240BatchSize500-16          10000            444120 ns/op            1712 B/op         17 allocs/op
BenchmarkClickhouseWrappedMultiline/Concurrency400BatchSize50-16           10000           1255278 ns/op            1758 B/op         16 allocs/op
BenchmarkClickhouseWrappedMultiline/Concurrency400BatchSize150-16          10000            429072 ns/op            1900 B/op         17 allocs/op
BenchmarkClickhouseWrappedMultiline/Concurrency400BatchSize500-16          10000            243843 ns/op            1641 B/op         18 allocs/op
BenchmarkCockroachDBDirectInsert/Concurrency80-16                          10000            119621 ns/op            1507 B/op         33 allocs/op
BenchmarkCockroachDBDirectInsert/Concurrency240-16                         10000            119200 ns/op            1506 B/op         33 allocs/op
BenchmarkCockroachDBDirectInsert/Concurrency400-16                         10000            114013 ns/op            1511 B/op         33 allocs/op
BenchmarkCockroachDBWrappedMultiline/Concurrency80BatchSize50-16           10000             42601 ns/op            1110 B/op         21 allocs/op
BenchmarkCockroachDBWrappedMultiline/Concurrency80BatchSize150-16          10000             35744 ns/op            1115 B/op         21 allocs/op
BenchmarkCockroachDBWrappedMultiline/Concurrency80BatchSize500-16          10000             43492 ns/op            1106 B/op         21 allocs/op
BenchmarkCockroachDBWrappedMultiline/Concurrency240BatchSize50-16          10000             39707 ns/op            1116 B/op         21 allocs/op
BenchmarkCockroachDBWrappedMultiline/Concurrency240BatchSize150-16         10000             21654 ns/op            1266 B/op         23 allocs/op
BenchmarkCockroachDBWrappedMultiline/Concurrency240BatchSize500-16         10000             42110 ns/op            1392 B/op         24 allocs/op
BenchmarkCockroachDBWrappedMultiline/Concurrency400BatchSize50-16          10000             37538 ns/op            1129 B/op         21 allocs/op
BenchmarkCockroachDBWrappedMultiline/Concurrency400BatchSize150-16         10000            105388 ns/op            1237 B/op         24 allocs/op
BenchmarkCockroachDBWrappedMultiline/Concurrency400BatchSize500-16         10000             23730 ns/op            1445 B/op         24 allocs/op
BenchmarkPostgresDirectInsert/Concurrency80-16                             10000             55056 ns/op            1542 B/op         34 allocs/op
BenchmarkPostgresDirectInsert/Concurrency240-16                            10000             54242 ns/op            1542 B/op         34 allocs/op
BenchmarkPostgresDirectInsert/Concurrency400-16                            10000             52326 ns/op            1544 B/op         34 allocs/op
BenchmarkPostgresWrappedCopyFrom/Concurrency80BatchSize50-16               10000             27918 ns/op            1013 B/op         25 allocs/op
BenchmarkPostgresWrappedCopyFrom/Concurrency80BatchSize150-16              10000             26767 ns/op            1016 B/op         25 allocs/op
BenchmarkPostgresWrappedCopyFrom/Concurrency80BatchSize500-16              10000             27998 ns/op             969 B/op         25 allocs/op
BenchmarkPostgresWrappedCopyFrom/Concurrency240BatchSize50-16              10000             22120 ns/op            1046 B/op         25 allocs/op
BenchmarkPostgresWrappedCopyFrom/Concurrency240BatchSize150-16             10000             12930 ns/op            1259 B/op         24 allocs/op
BenchmarkPostgresWrappedCopyFrom/Concurrency240BatchSize500-16             10000             12688 ns/op            1213 B/op         24 allocs/op
BenchmarkPostgresWrappedCopyFrom/Concurrency400BatchSize50-16              10000             23187 ns/op            1016 B/op         25 allocs/op
BenchmarkPostgresWrappedCopyFrom/Concurrency400BatchSize150-16             10000             11581 ns/op            1186 B/op         24 allocs/op
BenchmarkPostgresWrappedCopyFrom/Concurrency400BatchSize500-16             10000              9889 ns/op            1167 B/op         24 allocs/op
BenchmarkPostgresWrappedMultiline/Concurrency80BatchSize50-16              10000             17041 ns/op            1095 B/op         21 allocs/op
BenchmarkPostgresWrappedMultiline/Concurrency80BatchSize150-16             10000             16300 ns/op            1111 B/op         21 allocs/op
BenchmarkPostgresWrappedMultiline/Concurrency80BatchSize500-16             10000             16939 ns/op            1136 B/op         22 allocs/op
BenchmarkPostgresWrappedMultiline/Concurrency240BatchSize50-16             10000             15862 ns/op            1113 B/op         21 allocs/op
BenchmarkPostgresWrappedMultiline/Concurrency240BatchSize150-16            10000              9751 ns/op            1237 B/op         23 allocs/op
BenchmarkPostgresWrappedMultiline/Concurrency240BatchSize500-16            10000             10314 ns/op            1374 B/op         23 allocs/op
BenchmarkPostgresWrappedMultiline/Concurrency400BatchSize50-16             10000             14456 ns/op            1113 B/op         21 allocs/op
BenchmarkPostgresWrappedMultiline/Concurrency400BatchSize150-16            10000              9430 ns/op            1269 B/op         24 allocs/op
BenchmarkPostgresWrappedMultiline/Concurrency400BatchSize500-16            10000             11280 ns/op            1479 B/op         25 allocs/op
BenchmarkScyllaDBDirectInsert/Concurrency80-16                             10000             33374 ns/op            2647 B/op         40 allocs/op
BenchmarkScyllaDBDirectInsert/Concurrency240-16                            10000             23567 ns/op            2672 B/op         39 allocs/op
BenchmarkScyllaDBDirectInsert/Concurrency400-16                            10000             20770 ns/op            2690 B/op         38 allocs/op
BenchmarkScyllaDBWrappedBatch/Concurrency80BatchSize50-16                  10000             44820 ns/op            1025 B/op         12 allocs/op
BenchmarkScyllaDBWrappedBatch/Concurrency80BatchSize150-16                 10000             46413 ns/op            1045 B/op         12 allocs/op
BenchmarkScyllaDBWrappedBatch/Concurrency80BatchSize500-16                 10000             45582 ns/op            1031 B/op         12 allocs/op
BenchmarkScyllaDBWrappedBatch/Concurrency240BatchSize50-16                 10000             37751 ns/op            1035 B/op         12 allocs/op
BenchmarkScyllaDBWrappedBatch/Concurrency240BatchSize150-16                10000             25859 ns/op            1075 B/op         11 allocs/op
BenchmarkScyllaDBWrappedBatch/Concurrency240BatchSize500-16                10000             19145 ns/op            1053 B/op         11 allocs/op
BenchmarkScyllaDBWrappedBatch/Concurrency400BatchSize50-16                 10000             39768 ns/op            1039 B/op         12 allocs/op
BenchmarkScyllaDBWrappedBatch/Concurrency400BatchSize150-16                10000             16486 ns/op            1146 B/op         11 allocs/op
BenchmarkScyllaDBWrappedBatch/Concurrency400BatchSize500-16                10000             15578 ns/op            1054 B/op         11 allocs/op
PASS
ok      github.com/alexdyukov/syncinserter/v2   1914.129s
```

## License
MIT licensed. See the included LICENSE file for details.
