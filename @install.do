#!/bin/sh

DESTDIR=${DESTDIR:-/usr/local/bin/}
MANDIR=${MANDIR:-/usr/local/man/man1/}

install -t $DESTDIR bin/redux
xargs -L1 -i ln -sf $DESTDIR/redux $DESTDIR/{} < LINKS
xargs -L1 -i install -t $MANDIR doc/{}.1 < LINKS
