#! /usr/bin/env bash

dir="/tmp/kl/deployment-tracker"
mkdir -p "$dir"

pushd $dir

lockfile="./lock"

echo "tracking log" >./tracker.log
[ -f "$lockfile" ] && exit 0

touch $dir/check-online.lock
trap "rm -rf ./check-online.lock" SIGINT SIGTERM
while true; do
	[ -f "$lockfile" ] || exit 0
	timeout 1 ping -c 1 100.64.0.1 >>/tmp/kl/ping.stdout 2>>/tmp/kl/ping.stderr
	exit_code=$?
	echo "$(date +%T): exit code" >>/tmp/kl/ping.exit_code
	if [ $exit_code -eq 0 ]; then
		#echo "$(date +%T): online" > /tmp/kl/online.status
		echo "online" >>/tmp/kl/online.status
	else
		echo "offline" >>/tmp/kl/online.status
	fi
	sleep 2
done

popd
