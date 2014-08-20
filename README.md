# Description

redo is top down software build tool similar to make but simpler and more reliable.

It implements the concept somewhat sparsely described by [DJ Bernstein](http://cr.yp.to/redo.html).

# Overview

A redo based system must have a directory named .redo at the top level project
directory. 

It is created with the `redo-init` command, whose effects are idempotent.

The database contains file status and metadata. Its format is not specified,
but it must be capable of supporting multiple readers and multiple writers
(though not necessarily writing to the same sections). Redux implements a file based
database and a null database. The current file based database is probably about the
slowest. Fortunately, new databases can be easily plugged in.

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

Please see the individual command documentation for further details.
