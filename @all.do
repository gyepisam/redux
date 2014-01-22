#!/bin/sh
# This do script can be directly invoked, for bootstrapping with
# $  sh @all.do
# $  redo @all

set -e

# The go compiler does its own dependency tracking and is so fast that
# it is ironically simpler to always rebuild.
go build -o bin/redux redux/cmd

export DESTDIR=bin
export PATH=$(dirname $0)/bin:$PATH

# symlink to binary
xargs -L1 -i ln -sf redux bin/{} < LINKS

# create docs 
xargs -L1 -i bin/redux redo doc/{}.1 < LINKS
