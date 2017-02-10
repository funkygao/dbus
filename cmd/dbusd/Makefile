mock:build
	./dbusd -conf ../../etc/mock.cf

mysql:build
	./dbusd -conf ../../etc/mysql.cf

build:
	@go build -race
	@mkdir -p var
