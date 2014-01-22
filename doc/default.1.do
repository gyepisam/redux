
source=${2##redo-}.md

redo-ifchange $source
pandoc -s -t man $source
