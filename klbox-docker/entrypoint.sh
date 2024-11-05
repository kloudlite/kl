#! /usr/bin/env bash

set -o errexit
set -o pipefail

/start.sh

export SSH_PORT=$SSH_PORT

/usr/sbin/sshd -D -p "$SSH_PORT" &
pid=$!

cat >/kl-tmp/kill-sshd.sh <<EOF
sudo kill -9 $pid
EOF

wait $pid
