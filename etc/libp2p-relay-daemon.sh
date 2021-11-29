#!/bin/bash
set -e

# file descriptor limit for TCP connections; adjust according for your needs
ulimit -n 65536

# if you are running with custom config:
# exec /root/go/bin/relayd -id /root/relayd.identity -config /root/config.json

exec /root/go/bin/relayd -id /root/relayd.identity
