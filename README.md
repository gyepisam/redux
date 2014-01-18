## Description

redo is a simpler and more reliable make replacement tool.

The basic idea for it is, somewhat sparsely, described by [DJ Bernstein](http://cr.yp.to/redo.html).

I implemented a minimal set of redo tools some years ago and used them enough to become convinced
that the idea was worthwhile. However, they needed to be better, sharper, and faster before they could
replace Make in my toolbox. This set of tools meet that challenge ably and I hope they work as well for you.

## Installation

redo is written in Go and requires the Go compiler to be installed, either from (http://www.golang.com) or your favorite distribution channel. Having done that, the commands:

    $ go get github.com/gyepisam/redo
    $ go install github.com/gyepisam/redo

will fetch, build and install the binaries into $GOPATH/bin.

## Quick start

This package builds and installs four executable files:


    redo-init      -- run this on the command line to initialize a project directory
    redo           -- run this on the command line or in a .do script  
    redo-ifchange  -- run this in a .do script
    redo-ifcreate  -- run this in a .do script


For now, see the doc files in each executable's source directory.
