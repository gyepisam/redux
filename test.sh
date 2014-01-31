#!/bin/sh

set -e
sh $(dirname $0)/build
go test $@
sh redux/install-man-test.sh
