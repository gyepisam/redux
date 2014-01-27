b=$(basename $0)
b=${b#@}
b=${b%.do}
exec $(dirname $0)/$b
