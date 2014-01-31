#!/bin/sh

# Test redux installation.

set -e
scriptdir=$(realpath $(dirname $0))

workdir=$(mktemp --tmpdir --directory)

# avoid breaking inside the source dir
cd $workdir

trap "rm -rf $workdir" 0

expected=$(mktemp --tmpdir=$workdir redux-man-XXXXXX)

cat - > $expected <<EOS
redo.1
redo-ifchange.1
redo-ifcreate.1
redo-init.1
EOS

bindir=$workdir/bin
mkdir $bindir
bin=$bindir/redux

#build a binary
(cd $scriptdir && go build -o $bin)

# Run test and compares result to $expected
run () {
    dst=$1
    shift
    $@
    t=$(mktemp  --tmpdir=$workdir redux-man-mandir-XXXXXX)
    find $dst -type f -name '*.1' -printf '%f\n' | sort > $t
    cmp --verbose $expected $t
}

dir=$(mktemp --directory --tmpdir=$workdir redux-install-XXXXXX)

# 0 use --mandir flag value if it exists
run $dir/manA "$bin install --mandir $dir/manA man"

# 1 use MANDIR if it is set
run $dir/manB "env MANDIR=$dir/manB $bin install man"

# 2 use first non-empty path in MANPATH if it is set.
#   This is equivalent to: $(echo $MANPATH | cut -f1 -d:) except that we skip empty fields.
#   There are a few variations on MANPATH
run $dir/manC "env MANPATH=:$dir/manC $bin install man"
run $dir/manD "env MANPATH=$dir/manD: $bin install man"
run $dir/manE "env MANPATH=$dir/manE::does-not-exist $bin install man"

# 3 use $(dirname redux)/../man if it is writable

# does not exist yet, but will be created
run $bindir/../man "$bin install man"

# exists but empty
rm -rf $bindir/../man/*
run $bindir/../man "$bin install man"

# 4 use first non-empty path in `manpath` if there's one.
#   This is equivalent to: $(manpath 2>/dev/null | cut -f1 -d:) except that we skip empty fields.

# create a directory for manpath to return
alt_man_dir=$(mktemp --directory  --tmpdir=$workdir redux-test-4-XXXXXX)
echo MANPATH_MAP $bindir $alt_man_dir  > ~/.manpath

# make $bindir/../man a file so redux can't use it and, instead, calls manpath
rm -rf $bindir/../man
touch $bindir/../man

# manpath checks ~/.manpath entries against entries in $PATH, so we add it, temporarily.
OLDPATH=$PATH
PATH=$bindir:$PATH
run $alt_man_dir "$bin install man"
PATH=$OLDPATH

# 5 use '/usr/local/man'
# $bindir/../man is still unusable,
# now mock up a failing manpath
cat > $bindir/manpath  <<EOS
#!/bin/sh
exit 1
EOS
chmod +x $bindir/manpath

# Since we cannot write to /usr/local/man, there's no point in doing a comparison.
# Just look for an error.
PATH=$bindir:$PATH
# don't exit on error
message=$($bin install man 2>&1) || :
echo "$message" | egrep -q "^Error:.+error.+permission denied"
if test "$?" != "0" ; then
  echo "Expected error message: $message" > /dev/stderr
fi

# cleanup
rm -f  ~/.manpath
