test:
	go test -v -race

# This is only used by go 1.12
bench:
	go test -v -bench=. -run=none \
		-benchtime=300000x -benchmem \
		-memprofile mem.out -cpuprofile cpu.out .

install_graphviz:
	sudo apt install graphviz

cpu_pprof:
	go tool pprof -svg cpu.out > cpu.svg

mem_pprof:
	go tool pprof -svg mem.out > mem.svg