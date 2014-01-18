#!/bin/sh

redo @all
sudo install -t ${DESTDIR:-/usr/local/bin/} `sed s}^}bin/} < BIN`
