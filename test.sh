#!/bin/sh

set -e
d=$(dirname $0)

sh $d/build
$d/bin/redux install links
go test $@

sh redux/install-man-test.sh
