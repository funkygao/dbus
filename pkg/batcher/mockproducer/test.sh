#!/bin/sh

for i in `seq 1000`; do
    go run producer.go | tail -1 | grep passed
done
