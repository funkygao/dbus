#!/bin/sh
#=================================================================
# bring chaos monkey to dbusd cluster to test controller stability
#=================================================================
#
#------------
# how to use?
#------------
# open 2 terminals:
# term1$ make c2
# term2$ sh chaos.sh

for i in `seq 1000`; do
    ./dbusd -cluster -pprof :10120 -rpc 9877 -api 9897 &
    pid=$!
    sleep 10
    sleep $[ ( $RANDOM % 10 ) + 10 ]s
    echo "killing $pid"
    kill $pid
    sleep $[ ( $RANDOM % 10 ) + 15 ]s
done

