mock:build
	./dbusd -conf ../../etc/mock.cf -routerstat=false

batcher:build
	./dbusd -conf ../../etc/batcher.cf -routerstat=false

v2:buildv2
	./dbusd -conf ../../etc/mock.cf -routerstat=false

stream:buildrace
	./dbusd -conf ../../etc/stream.cf -routerstat=false

inputs:build
	./dbusd -conf ../../etc/inputs.cf -routerstat=false

fanout:buildrace
	./dbusd -conf ../../etc/fanout.cf

db1:
	for i in `seq 1000`;do echo "insert into t1 values(39)" | mysql -uroot db1 ; done

db2:
	for i in `seq 1000`;do echo "insert into t2 values(86)" | mysql -uroot db2 ; done

dispatch:buildrace
	./dbusd -conf ../../etc/dispatch.cf

complex:buildrace
	./dbusd -conf ../../etc/complex.cf

resource:
	dbc resources -add in.binlog-mysql:local://root@localhost:3306

resources:
	dbc resources -add in.test-mysql:local://root@localhost:3306/test
	dbc resources -add in.mysql-mysql:local://root@localhost:3306/mysql

zkconf:
	@cat ../../etc/canal_many.cf | zk set -p /dbus/conf

sub:build
	@cat ../../etc/sub.cf | zk set -p /dbus/conf
	-dbc resources -add in.kafka-kafka:local://me/foobar#0
	UPGRADE_ENDPOINT=http://localhost/dbusd ./dbusd -cluster

canal:build zkconf resources
	-dbc resources -add in.binlog-mysql:local://root@localhost:3306
	UPGRADE_ENDPOINT=http://localhost/dbusd ./dbusd -cluster

c1:build zkconf
	UPGRADE_ENDPOINT=http://localhost/dbusd ./dbusd -cluster -pprof :10120 -rpc 9877 -api 9897 -loglevel trace

c2:build zkconf
	UPGRADE_ENDPOINT=http://localhost/dbusd ./dbusd -cluster -pprof :10121 -rpc 9878 -api 9898 -loglevel trace

c3:build zkconf
	UPGRADE_ENDPOINT=http://localhost/dbusd ./dbusd -cluster -pprof :10122 -rpc 9879 -api 9899 -loglevel trace

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
