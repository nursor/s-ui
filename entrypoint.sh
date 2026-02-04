#!/bin/sh
# Start cron daemon in background
crond -b -l 8

nginx -g 'daemon off;' &
./sui migrate
./sui
