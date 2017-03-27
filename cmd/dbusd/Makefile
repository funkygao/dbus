mock:
	@go build
	./dbusd -conf ../../etc/mock.cf -routerstat=false

batcher:
	@go build
	./dbusd -conf ../../etc/batcher.cf -routerstat=false

v2:buildv2
	./dbusd -conf ../../etc/mock.cf -routerstat=false

inputs:
	@go build
	./dbusd -conf ../../etc/inputs.cf -routerstat=false

fanout:build
	./dbusd -conf ../../etc/fanout.cf

dispatch:build
	./dbusd -conf ../../etc/dispatch.cf

complex:build
	./dbusd -conf ../../etc/complex.cf

canal:
	@go build
	@cat ../../etc/canal.cf | zk set -p /dbus/conf
	./dbusd -cluster

dump:build
	./dbusd -conf ../../etc/complex.cf -dump canal.png

build:
	@go build -race
	@mkdir -p var

buildv2:
	go build -tags v2

pprof:
	go tool pprof http://localhost:10120/debug/pprof/profile?seconds=60

flame:
	@go-torch -u http://localhost:10120/debug/pprof/profile --title=dbusd
	@open torch.svg
