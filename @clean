#!/bin/sh
rm -f $(sed 's}^}bin/}' < BIN) $(find $(dirname $0) -path ./t -prune -o -type f -name '*~')
