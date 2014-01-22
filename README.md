# Description

redux is top down software build tool similar to make but simpler and more reliable.

It implements and extends the redo concept somewhat sparsely described by [DJ Bernstein](http://cr.yp.to/redo.html).

I implemented a minimal set of redo tools some years ago and used them enough to become convinced
that the idea was worthwhile. However, they needed to be better, sharper, and faster before they could
replace Make in my toolbox, thus, this tool.

# Installation

redux is written in Go and requires the Go compiler to be installed.
Go can be fetched from the [Go site](http://www.golang.com) or your favorite distribution channel.

Assuming Go is installed, the commands:

    $ go get github.com/gyepisam/redux
    $ cd $GO_PATH/github.com/gyepisam/redux
    $ sh @all.do 
    $ sudo bin/redux  @install

will fetch, build and install redux to the default location of /usr/local/bin, with man
pages in /usr/local/man/man1.

    $ export DESTDIR=/some/other/path/bin MANDIR=/some/other/path/man/man1
    $ sudo bin/redux @install 

@install installs a single multi-call binary named redux, along
with symlinks for the following commands

       init -- Creates or reinitializes one or more redo root directories.

   ifchange -- Creates dependency on targets and ensure that targets are up to date.

   ifcreate -- Creates dependency on non-existence of targets.

       redo -- Builds files atomically.

Each symlink, except redo, is prefixed with redo. So the init command can be invoked as
`redux init` or as `redo-init`. The same pattern holds true for the other commands, except
for redo, which is `redux redo` or plain `redo`.

Each command is further documented in the [doc directory](/doc).

# Overview

A redo based system must have a directory named .redo at the top level project
directory. 

It is created with the `redo-init` command, whose effects are idempotent.

The database contains file status and metadata. Its format is not specified,
but it must be capable of supporting multiple readers and multiple writers
(though not necessarily writing to the same sections). Redux implements a file based
database and a null database. The current file based database is probably about the
slowest. Fortunately, new databases can be easily pluggled in.

In the redo system, there are three kinds of files.

* `do` files are scripts that build targets. In addition, they establish
file dependencies by calling `redo-ifcreate` and `redo-ifchange`. `do` files
are invoked by the `redo` commands directly and indirectly by `redo-ifchange`.

* `target` files are generated or otherwise composed, likely from other files, by `do` scripts

* `static` or `source` files are manually created/edited. They are tracked
for changes, but are not generated. As such, they have no do files.

There are two kinds of dependencies

* ifchange denotes a relationship where a change in one file marks the other as out of date.
  A ifchange B implies that a change in B invalidates A. Change includes creation, modification or deletion.

* ifcreate denotes a dependency on the creation or deletion of a file.
  A ifcreate B implies that when B comes into existence or ceases to exist, A will be outdated.
