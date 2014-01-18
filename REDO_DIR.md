* The .redo directory

The .redo directory marks the project root directory and holds configuration and data files.

.redo/data is a directory

entries are directories whose names are sha1 checksums of file paths relative to the .redo directory.
[TODO] Full paths would be better for cross project references, but would invalidate all references if the
project directory is moved.

Each entry has three contents

meta is a file containing the following data: 

 path
 size
 timestamp
 sha1sum

dependents and prerequisites are directories

In the dependency graph

 A ifchange X.ext
 B ifcreate X.ext
 X.ext ifchange D
 X.ext ifcreate E

Where default.ext.do is the most relevant do file:

The metadata for X.ext would look like:

# file metadata
.redo/data/checksum(X.ext)/METADATA

# rebuild flag
.redo/data/checksum(X.ext)/REBUILD

# dependents to be outdated
.redo/data/checksum(X.ext)/outdates/ifchange/checksum(A)
.redo/data/checksum(X.ext)/outdates/ifcreate/checksum(B)

# prerequisites
.redo/data/checksum(X.ext)/requires/ifchange/checksum(D)
.redo/data/checksum(X.ext)/requires/ifchange/checksum(E)
.redo/data/checksum(X.ext)/requires/auto/ifcreate/checksum(default.ext.do)
.redo/data/checksum(X.ext)/requires/auto/ifchange/checksum(X.ext.do)

* Suppressing target output

`.do` files whose names are prefixed with '@' are run for their side effects
and not for generating output. Redo does not create a temporary file when running such
files and uses '/dev/stdout' as their output file name.

A call to redo without an argument will search for a file named `@all.do`
in the current directory.

# Project directory

A redo based system must have a directory named .redo at the top level project
directory. 

It is created with the `redo-init` command, whose effects are idempotent.

The database contains file status and metadata. Its format is not specified,
but it must be capable of supporting multiple readers and multiple writers
(though not necessarily writing to the same sections).

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


