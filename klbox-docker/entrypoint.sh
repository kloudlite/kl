#!/bin/bash
tail -f /dev/null

ls -al /tmp

(tail -f /tmp/stdout.log) &
pid=$!
(tail -f /tmp/stderr.log) &
pid="$pid $!"

trap "eval kill -9 $pid" EXIT TERM
/start.sh $@ > /tmp/stdout.log 2> /tmp/stderr.log
