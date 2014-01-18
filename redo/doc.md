## The redo command

redo is an incremental build tool.

# Usage

    redo [OPTIONS] [TARGETS]

# Options


    -help, -h   Show help
    -verbose, -v Run verbosely. This option can be repeated for intensity.
    -trace, -v   Trace shell script execution
    -task        Run the targets as task scripts.

redo normally requires one or more target arguments.
If no target arguments are provided, redo searches for the script '@all.do' in
the current directory, and runs it if found.

redo searches for the target's build script [See] (#do-search) and, when it finds it,
runs the do file with the current working directory set to the do file's directory.

In the case where the do file exists, it is assumed to be an sh script and called with three arguments

$1 = path to target 
$2 = basename of target without a suffix
$3 = temporary file name

The script is called with stdout opened to a temporary file (which is unnamed and different
from $3) and the script is expected to produce output on stdout or write to $3. It is an error to do both.

If the script completes successfuly, redo renames the temporary file to the target file and updates its
database with the new file's metadata record. Since only one of the two temporary files can have content,
redo has no trouble selecting the correct one. Conversely, if neither file has content, then
either is a valid candidate.

In the do file, which is an sh script, a call to 

    redo-ifchange A B C

specifies A, B, and C as prerequisites for the target file.

Similarly, a call to 

    redo-ifcreate A

specifies that the target should be rebuilt when the file A first appears or is deleted.


# .do file

A target is built with a `.do` file in the same or a higher level directory
(up to the project directory). For a given target named target.ext,
the corresponding .do file may be named, in order of decreasing specificity,
target.ext.do, default.ext.do or default.do. A do file is normally expected
to produce the named target file either on stdout or to the file specified by
its parameter $3.. It is an error for a do file to write to both outputs.

The do file is executed with /bin/sh in the do file's directory.

As a special case, a do file whose name is prefixed with '@' is run for
side effect and is not expected to produce a target. This is analogous
to make's `.PHONY` targets. Any do file can also be run as a task by
invoking 'redo' with the '-task' flag.

# .do file search

The most specific file is expected to exist in the directory as the target.
The other two are searched for in the target's directory and parent directories
up to the `.redo` directory parent. The `.redo` directory marks the project
directory and allows multiple redo based projects to co-exist at different levels
of a directory tree.

target is a filename with the extension EXT. Redo searches for a `.do` file, a script whose execution
will produce the desired target file. There are several possible names for the script, from most specific
to least, and the appropriate file could exist anywhere between the target's directory and the project top
level. This flexibility makes it possible to use the same script to generate multiple targets, if necessary.

Starting in the directory containing target.EXT, redo searches for one of the files
target.EXT.do, default.EXT.do, and default.do. If none is found, redo repeats the search sequentially
in parent directories.  The search stops when a script file is found or
redo reaches the the project root, which is denoted by the directory `.redo`, created by `redo-init`.

The first correct `.do` file found is used to produce target.ext.



## Example 

To generate a file A based on a dependency on B and C, create the file A.do

   cat > A.do <<EOS
   redo-ifchange B C
   echo 'A Contents'
   EOS

then run the command 

    $ redo A

