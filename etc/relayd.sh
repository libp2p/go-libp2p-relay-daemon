#!/bin/bash
set -e

# set ulimit for file descriptors; only necessary if you are using TCP
# ulimit -n 1048576

# if you are running with custom config:
# exec /root/go/bin/relayd -id /root/relayd.identity -config /root/config.json

exec /root/go/bin/relayd -id /root/relayd.identity
