#!/bin/sh

find $(dirname $0) -path ./t -prune -o -type f -name '*~' -exec rm -f {} \;
