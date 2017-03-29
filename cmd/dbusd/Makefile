mock:build
	./dbusd -conf ../../etc/mock.cf -routerstat=false

batcher:build
	./dbusd -conf ../../etc/batcher.cf -routerstat=false

v2:buildv2
	./dbusd -conf ../../etc/mock.cf -routerstat=false

inputs:build
	./dbusd -conf ../../etc/inputs.cf -routerstat=false

fanout:buildrace
	./dbusd -conf ../../etc/fanout.cf

dispatch:buildrace
	./dbusd -conf ../../etc/dispatch.cf

complex:buildrace
	./dbusd -conf ../../etc/complex.cf

canal:build
	@cat ../../etc/canal.cf | zk set -p /dbus/conf
	./dbusd -cluster

c1:build
	./dbusd -cluster -pprof :10120 -rpc 9877 -api 9897

c2:build
	./dbusd -cluster -pprof :10121 -rpc 9878 -api 9898

dump:buildrace
	./dbusd -conf ../../etc/complex.cf -dump canal.png

build:
	@go build

buildrace:
	@go build -race

buildv2:
	go build -tags v2

pprof:
	go tool pprof http://localhost:10120/debug/pprof/profile?seconds=60

flame:
	@go-torch -u http://localhost:10120/debug/pprof/profile --title=dbusd
	@open torch.svg
