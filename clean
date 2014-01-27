#!/bin/sh
rm -f $(sed 's}^}bin/}' < LINKS) $(find $(dirname $0) -path ./t -prune -o -type f -name '*~')
