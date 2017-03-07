mock:
	@go build
	./dbusd -conf ../../etc/mock.cf

mysql:build
	./dbusd -conf ../../etc/mysql.cf

build:
	@go build -race
	@mkdir -p var

pprof:
	go tool pprof http://localhost:10120/debug/pprof/profile?seconds=60
