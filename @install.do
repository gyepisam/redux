#!/bin/sh

sh @all.do
sudo install -t ${DESTDIR:-/usr/local/bin/} `sed s}^}bin/} < BIN`
