## Description

redo is a simpler and more reliable make replacement tool.

The basic idea for it is, somewhat sparsely, described by Dan Bernstein here (http://cr.yp.to/redo.html).

## Operation

In the the command 

    $ redo target.EXT

target is a filename with the extension EXT. Redo searches for a `.do` file, a script whose execution
will produce the desired target file.

Starting in the directory containing target.EXT, redo searches for one of the files
target.EXT.do, default.EXT.do, and default.do. If none is found, redo searches sequentially
in parent directories.  The search stops when a script file is found or
redo reaches the the project root, which is denoted by the directory `.redo`, normally created by `redo-init`.

The first correct `.do` file found is used to produce target.ext.

In the case where the do file exists, it is assumed to be an sh script and called with three arguments

$1 = full path to target 
$2 = basename of target without a suffix
$3 = temporary file name

The script is called with stdout opened to a temporary file (which is unnamed and different
from $3) and the script is expected to produce output on stdout or write to $3. It is an error to do both.

If the script completes successfuly, redo renames the tmp file to the target file and updates its
database with the new file's metadata record. Since only one of the two temporary files can have content,
redo has no trouble selecting the correct one. Conversely, if neither file has content, then
either is a valid candidate.

In the do file, which is an sh script, a call to 

    redo-ifchange A B C

specifies A, B, and C as prerequisites for the target file.

Similarly, a call to 

    redo-ifcreate A

specifies that the target should be rebuilt when the file A first appears or is deleted.

## Usage

To generate a file A based on a dependency on B and C, create the file A.do

   cat > A.do <<EOS
   redo-ifchange B C
   echo 'A Contents'
   EOS

then run the command 'redo A'


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

# .do file

A target is built with a `.do` file in the same or a higher level directory
(up to the project directory). For a given target named target.ext,
the corresponding .do file may be named, in order of decreasing specificity,
target.ext.do, default.ext.do or default.do. A do file is normally expected
to produce the named target file either on stdout or to the file specified by
its parameter $3. It is an error for a do file to write to both outputs.

The do file is executed with /bin/sh with PWD set to the do file's directory.

As a special case, a do file whose name is prefixed with '@' is run for
side effect and is not expected to produce a target. This is analogous
to make's `.PHONY` targets.

# .do file search

The most specific file is expected to exist in the directory as the target.
The other two are searched for in the target's directory and parent directories
up to the `.redo` directory parent. The `.redo` directory marks the project
directory and allows multiple redo based projects to co-exist at different levels
of a directory tree.

# The redo-init command

The redo-init command creates and initializes a .redo directory.
The .redo directory holds the redo configuration and database files
and also demarcates the root of the redo enabled project.

usage:
        
        redo-init DIRECTORY
        env REDO_DIR=DIRECTORY redo-init
        redo-init

        The target directory for the initializing can be specified as an argument
        as in case 1, provided by the REDO_DIR environment value as in case 2, or
        not provided at all, as in case 3, in which event, the current directory
        is used.

# The redo command

A redo invocation requires an explicit target or an implicit target of @all.

The `redo` runs a target's do file with the current working directory set to the
do file's directory and the environment containing the variable `REDO-TARGET=target`

The following pseudo code illustrates how redo works

         
    procedure outdate-dependents(dependency-type, target)
        for each dependent of type dependency-type on target
            mark dependent as out of date
            
    procedure add-file-record(file do-file timestamp hash)
        # note: do-file is nil when file is source, not target
        add record

    procedure update-file-record(file do-file timestamp hash)
        update record
           
    procedure delete-file-record(file)
        delete file record
        [call procedure outdate-dependents 'ifcreate' file]
        delete dependency records to file
        delete dependency records from file
        
        
    procedure run-do-file(target do-file)
        run do-file [1]
        # check for multiple outputs...
        
        created = record.hash == nil
        if result.hash != record.hash
           save new file
           update target record(target do-file exit-status content-hash timestamp)
           if created
              [call procedure outdate-dependents 'ifcreate' file]
              
           [call procedure outdate-dependents 'ifchange' file]
           
    procedure file-timestamp-changed (record file)
       file.timestamp != record.timestamp
       
    procedure file-hash-changed(record file)
       file.hash != record.hash
       
    procedure file-changed?(record file)
       file-timestamp-changed?(record file) or file-hash-changed(record file)


    procedure redo-target(target, do-file, do-files-not-found)
              [call procedure redo-ifchange(target, do-file)]
              for each file in do-files-not-found
                  [call procedure redo-ifcreate(target, file)]
              [call procedure run-do-file target do-file]
              
    procedure is-doable?(record)
              record.do_file != nil
              
    procedure redo(target file)
      do-file, do-files-not-found = find-do-file(target)
      record, record-found? = find-file-record(target)
      
      if target file exists
          if record-found?
              if do-file
                  [call procedure redo-target target do-file do-files-not-found]
              else if is-doable?(record)
                  error "Missing do file for target" (note: 3)
              else if file-changed?(record target) #target is a source file
                   [call procedure update-file-record target]
                   [call procedure outdate-dependents 'ifchange' target]
          else
              if do-file
                  [call procedure redo-target target do-file do-files-not-found]
              else #target is source
                  [call procedure add-file-record target]
                  [call procedure outdate-dependents 'ifcreate' target]
              
      else # target file does not exist
          if record-found?
              # existed at one point but was deleted
              [call procedure outdate-dependents 'ifcreate' file]
              if do-file
                 [call procedure redo-target target do-file do-files-not-found]
              else if is-doable?(record)
                 error "cannot find do file for target"
              else # target is a deleted source file
                 [call procedure outdate-dependents 'ifchange' target]          
                 [call procedure forget-file target]
                 error "source file does not exist"
          else
              if do-file
                 [call procedure redo-target target do-file do-files-not-found]
              else
                error "cannot redo unknown file: target file"
            
Notes:

[1] In the special case where a class of targets are generated by a default.do file,
but some files are 'source' files and not generated, it is necessary for the default.do
file to recognize the special case and 'generate' the file with cat.

[2] If a more specific do file is created for a file, 
the current redo-ifcreate rule (for the less specific file) would be invoked once.
but the rules should then be changed to 'redo-ifchange do file' for the more specific
This can be accomplished by deleting the old set and generating a new set.

[3] This implies a need for the command '`redo-forget` target' to remove a target's record from the
database.

[4] In this implementation, the case where a more specific do file is added and the less specific
one modified will cause both scripts to mark the target as out of date.


# redo-static

Static or source files represent leaf nodes in the dependency graph.
They do not have dependencies on other files.  However, since they are
edited, they do change and must be tracked in order to trigger dependency
actions. When 

  procedure redo-static(target)
            [call procedure update-file-record record target]
            [call procedure outdate-dependents "if-change" target]
                 
# redo-ifchange

The `redo-ifchange` command is used in a do file. When the do file for a target A contains the
line `redo-ifchange B`, this means that the target A depends on B and A should be rebuilt if B
changes. Note that since redo-ifchange is invoked in the context of a do file, the environment variable
'REDO-TARGET', denoting the current target, exists.

Conceptually, redo-ifchange performs two tasks.

First, it ensures that  a dependency record of type "A ifchange B' is created so that
a change in B immediately outdates A.

Second, if B is out of date, redo-ifchange ensures that B is made up to date.

B is considered out of date if it does not exist, is not in the database, is flagged as out of date, 
has been modified or any of its dependents are out of date. Obviously this process
may recurse.

   procedure out-of-date?(target)

       if not file-exists? target
          return true

       record, record-found? = find-file-record(target)
       if not record-found?
          return true

       if record.outdated?
          return true
          
       if file-changed?(record, target)
          return true
             
       if any? (dependencies file) changed?
             return true

          
   procedure redo-ifchange(target)

       ensure-dependency-record((get-env "REDO-TARGET") "ifchange" target)
       
       if out-of-date?(target)
          do-file, do-files-not-found = find-do-file(target)
          if do-file
             [call procedure run-do-file target do-file]
          else
            record = find-file-record(target)
            if is-doable?(record)
                 error "Missing do file for target"
            else if file-changed(record, target)
                 [call procedure redo-static target]


Notes:
[1] This means that unchanging 'source' files will always be considered to be out of date
only to be left unchanged by the `redo` algorithm. While this seems wasteful, it does serve
the purpose of allowing the redo algorithm to detect when a source file is changed to a generated
one.  

# redo-ifcreate

The `redo-ifcreate` command is used to mark the current target as out of date when the non-existent dependency
file comes into existence or is deleted.
The command `redo-ifcreate B C D`, in A.do, would execute the following pseudo code.

    for each dependency in B C D
       ensure dependency([env get REDO-TARGET], ifcreate, $dependency) 


# redo-forget

The `redo-forget` command is used to clear database records for targets that have been deleted.
It is invoked as `redo-forget TARGET...`
