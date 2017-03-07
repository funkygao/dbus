bench:
	@go test -v -benchmem -benchtime=30s -test.bench="BenchmarkBatcherWriteWithBatchSize100" -cpuprofile cpu.pprof
	@go tool pprof batcher.test cpu.pprof

clean:
	rm -f batcher.test cpu.pprof
