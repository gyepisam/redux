#!/bin/sh

# run a few tests on redo-init

prog=$1


if [ -z "$prog" ] ; then
    f=$(dirname $0)/redo-init
    if test -x $f ; then
      prog=$f
    fi
fi

if [ -z "$prog" ] ; then
    echo "usage $0 path/to/redo-init" > /dev/stderr
    exit 1
fi

prog=$(readlink -f $prog)

set -e

arg_test () {
    cmd=$1
    d=$(mktemp --directory)
    p=project
    mkdir $d/$p
    $cmd $d $p 
    test -d $d/$p/.redo || echo "failed $cmd"
    rm -r $d
}

with_cwd () {
    cd $1/$2 && $prog
}

with_cwd_env () {
    cd $1 && env REDO_DIR=$2 $prog
}

with_cwd_arg () {
    cd $1 && $prog $2
}

with_fullpath_env () {
    env REDO_DIR=$1/$2 $prog
}

with_fullpath_arg () {
    $prog $1/$2
}

test_with_multiple_args () {
    a=$(mktemp --directory)
    b=$(mktemp --directory)
    c=$(mktemp --directory)

    $prog $a $b $c

    test -d $a/.redo || echo "failed test_with_multiple_args a"
    test -d $b/.redo || echo "failed test_with_multiple_args b"
    test -d $c/.redo || echo "failed test_with_multiple_args c"

    rm -r $a $b $c
}

arg_test with_cwd
arg_test with_cwd_env
arg_test with_cwd_arg
arg_test with_fullpath_env
arg_test with_fullpath_arg
test_with_multiple_args
