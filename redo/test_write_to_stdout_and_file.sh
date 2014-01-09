
base=$(realpath $(dirname $0)/..)
d=$(mktemp --directory)
$base/redo-init/redo-init $d
cat <<EOS > $d/both.do
echo "writing to stdout!"
echo "writing to file too" > \$3
EOS

err=$(mktemp)
$base/redo/redo $d/both 2> $err
e=$?
n=1
if [ $e -ne $n ] ; then
  echo "Expected exist status $n, not $e" > /dev/stderr
fi

if ! egrep -q '^Error:.+to stdout and to file' $err ; then
 echo "Expected error message" > /dev/stderr
fi

rm -rf $err $d
