#!/bin/sh

set -e

while read name ; do
  go build -o bin/$name $name/main.go
done < BIN
