#!/bin/sh

set -e

tests="$@"

base=$(realpath $(dirname $0)/..)
PATH=$base/redo:$base/redo-init:$PATH

datafile=$(mktemp /tmp/redo-data-XXXXXXXXXX)
outfile=$(mktemp /tmp/redo-output-XXXXXXXXXX)

trap "rm -f $datafile $outfile" 0

target=peach

run_test () {
    s=$(pwd)
    d=$(mktemp --directory /tmp/redo-XXXXXXXXXX)
    redo-init $d
    content=$(mktemp /tmp/redo-content-XXXXXXXXXX)
    random-text > $content
    (
      echo "cat <<EOS"
      cat $content
      echo "EOS"
    ) > $d/$target.do
    $1 $d
    find $d/.redo/data -type f -printf '%P\n' >> $datafile
    f=$d/$target
    test -f || echo "Failed to produce target: $f" > /dev/stderr
    echo $f >> $outfile
    cat $f >> $outfile
    echo >> $outfile

    cmp -b $content $f || echo "Diff failed: $f" > /dev/stderr
    cd $s
    rm -rf $d $content
}

test_pwd_file () {
    cd $1
    redo $target
}

test_pwd_relative_file () {
    cd $1
    redo ./$target
}


test_pwd_relative_path () {
  cd $(dirname $1) 
  redo $(basename $1)/$target
}

test_relative_path () {
    cd $1/..
    redo ./$(basename $1)/$target
}


test_indir_fullpath () {
    cd $1
    redo $1/$target
}

test_outdir_fullpath () {
    cd /tmp
    redo  $1/$target
}

if [ -z "$tests" ] ; then
  tests=$(awk '$1 ~ /^test_/ && $2 == "()" {print $1}' $0)
fi

test_count=0
for name in $tests ; do
  run_test $name
  test_count=$(expr $test_count + 1)
done

data=$(wc -l < $datafile)
data_count=3 # metadata, .do metadata, requires
data_exp=$(expr $test_count \* $data_count)
failed=0
if [ $data -ne $data_exp ] ; then
  (
  echo "expecting $data_exp lines of data but got $data"
  cat $datafile
  ) > /dev/stderr
  failed=1
else
  data_uniq=$(sort -u $datafile | wc -l)
  if [ $data_uniq -ne $data_count ] ; then
    echo "expecting $n unique lines of data but got $data_uniq" > /dev/stderr
    failed=1
  fi
fi

if test $failed = 1 ; then
   trap "" 0
   echo "see $datafile and $outfile"
fi
