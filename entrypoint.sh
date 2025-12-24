#!/bin/sh
nginx -g 'daemon off;' &
./sui migrate
./sui