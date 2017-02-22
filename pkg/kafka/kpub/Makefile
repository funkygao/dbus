all:
	go run kpub.go -z local -c me -t foobar -ack all -sync=false -sleep 1s -sz 20

race:
	go run -race kpub.go -z local -c me -t foobar -ack all -sync=false -sleep 5s -sz 20 -s=false

bench:
	go run kpub.go -z local -c me -t foobar -ack all -sync=false -sz 2500
