# v1.4.0

```
go test -v  -bench=. -run=none \
        -benchtime=300000x -benchmem \
        -memprofile mem.out -cpuprofile cpu.out .
goos: linux
goarch: amd64
pkg: github.com/kelseyhightower/envconfig
BenchmarkGatherInfo-4             300000             41525 ns/op           17344 B/op        270 allocs/op
PASS
ok      github.com/kelseyhightower/envconfig    12.652s
```
