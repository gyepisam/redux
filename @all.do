#!/bin/sh
# This do script can be directly invoked, for bootstrapping with
# $  sh @all.do
# $  redo all

# The go compiler does its own dependency tracking and is so fast that
# it is ironically simpler to always rebuild.
xargs -P 0 -L1 -i  go build -o bin/{} {}/main.go < BIN
