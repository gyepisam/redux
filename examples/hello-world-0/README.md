<A name="toc1-0" title="A hello world example for redux" />
# A hello world example for redux

Redux is a top down software build tool. The best way to explain it is to build something with it.
In this example, we are going to build a `hello world` program in C. We'll assume that you have read the
documentation for redux and know some C, but the example should be useful in any case. If you are reading
this from a source directory, you should have all the accompanying files and can play with them. If you are
not, the source files are available [here](https://github.com/gyepisam/blob/master/doc/examples/hello-world-0)


Our basic C hello world program, which consists of a couple of headers and a function, looks like this:

    /* This program prints a greeting */
    
    #include <stdio.h>
    #include <stdlib.h>
    
    int main(int argc, char *argv[]) {
      fprintf(stdout, "hello world\n");
      exit(0);
    }

The usual command to build this program is

    $ cc -o hello hello.c

Which you normally type at the command line.

However, redux is a software build tool so we'll automate the process. First, before we can use redux in a project,
we have to tell redux that we are doing so. The command

    $ redo-init

Initializes the project directory by creating a '.redo' directory in it. This command only needs
to be run once. It does no harm if you run it multiple times, and redux will remind you if you
forget to run it.

Now that we've initialized, we can continue with the automation.
In redux, the script to create a target `x` is called `x.do` so we'll create the file `hello.do`,
which contains almost the same command as before.

    redo-ifchange $2.c
    cc -o $3 $2.c

It is worth explaining why this looks different.
First, we've added the line `redo-ifchange $2.c`, which says that if hello.c changes,
hello needs to be rebuilt. In this case we are asking redo-ifchange to note that hello.c
is a prerequisite for hello, which is what we want.

redux executes the redo script with three arguments.

      * Argument $1 contains the name of the `target` which was given to redo.

        If you say `redo xyz.txt`, the $1 argument is xyz.txt, assuming, of course that there is a build script for it.

      * Argument $2  is the same as $1 but without the extension.
        In our example, it would be 'xyz'.

      * The $3 argument contains the name of a temporary file that redux creates to hold the output of of the script.
        So, in our script here, the '$2' argument contains the name of the target, without its extension,
        which is 'hello'. To this, we append a '.c' to create the name 'hello.c' and we are back at our source file.

So, knowing all that, we can see that the second line invokes the c compiler on hello.c ($2.c) and sends the output to the
temporary file $3. If the command completes successfuly, redo will move the temporary output to the correct name.
In this case, we have to specify an output path to the compiler because we don't normally spew compiler output to stdout.
We could, but we don't so we send it to a file instead. However, many other programs happily send their output to stdout,
and in that instance, they do not need an output specification when used in a do script. redux will redirect their output
to the correct location. While on the topic, it is important to note that a script cannot write to stdout *and* to the $3 file.
That is an error and would make redux complain. This restriction exists so redux can determine which output represents the
target.

Also, it is important to remember that do scripts *are* shell scripts. In fact, they are run with /bin/sh.
As such, you can use any shell construct you find necessary to get the job done. We could just as easily have written

    redo-ifchange hello.c
    cc -o $3 hello.c

However, redo scripts are reusable, and using variable names makes it possible
to do so.  BTW, the argument the $3 argument is necessary because redux guarantees that the target
file is never left in an incomplete state and uses temporary files to implement the guarantee.
However, this means that the do script must either write to stdout *or* to $3. Yes, that was worth repeating.

Now, if you type

     $ redo hello

redux will run the do script and create the file for you. If you had, instead, used the command

     $ redo -v hello

redux would have emitted some output in the process. In general, redux is quiet and unobtrusive.
It will do its job quietly and only speak up when asked or when something goes wrong. If you don't
see any error messages, it means there were none.

Now, you can run the hello program to see the expected output.

     $ ./hello

We can automate even this step! In redux, there are two kinds of do scripts.
Regular do scripts are expected to generate output. Task scripts are not, and
are, instead, run for their side effects.

So we can create a script to run hello world program for us.
A task script is usually named with an '@' prefix, so our file is named '@hello.do' and contains:

    bin=${1#@}
    redo-ifchange $bin
    ./$bin

This file is worth explaining. First, the line

     bin=${1#@}

is an sh construct that says, set the variable bin to the value of argument $1, but without the '@' prefix.
The '#' is part of a small family of shell substitution commands, all of which are worth knowing.

The prefix exists in the argument because our do file is called @hello.do and will be invoked with the command
'redo @hello' at which point, redo will pass the target, @hello, as argument $1.

We assign the value to bin because it is used twice. On general principle, it makes sense to save the result
of a computation and refer to it rather than repeating the computation. In the first reference, we add a dependency
on the target and in the second, we run it. Now, if we say

    $ redo @hello

We see the expected output again. Note the chain of dependencies: the target @hello depends on hello,
which itself depends on hello.c. The chain is built incrementally, but redux is able to trace them.
If we delete the 'hello' file or edit the hello.c file, we can actually see redux following the dependencies
all the way through. With the make program, you could touch a file to mark it as changed. However, redux uses the file
content and not the  modification time as a change indicator so that wouldn't work.

Obviously, we can do that with a simple rm command. However, this presents an opportunity to create the @clean target.
This target's purpose is to cleanup. If you know make, this would be equivalent to the 'clean' dependency in make.
The file contains the usual clean incantations; we want to delete the file we build, along with any editor flotsam.

    rm -f hello *~ 

Now, we can say

     $ redo @clean

Next, we say

     $ redo -v @hello

and see that redux first built the hello file before running the @hello target to produce the greeting. BTW, the name @hello.do
is not special, we could just as well have named it @greeting.do. However, the name hello.do *is* special because it produces
the hello file. Because task scripts are not expected to produce output, they can be named anything. Regular scripts, however,
must be named after their output file.

To recap, in this example, we created a simple hello world project that uses redux in order to show the basic steps of process
and used redo-init, redo-ifchange and redo. The final command, redo-ifcreate, is not as common and can be left off until later.
The only other useful item to know about are default do files. However, that is covered in the redo documentation and I really
want you to read it, so I'll point you [there](https://github.com/gyepisam/redux/blob/master/doc/redo.md)







