#!/bin/sh

install -t ${DESTDIR:-/usr/local/bin/} `sed s}^}bin/} < BIN`
