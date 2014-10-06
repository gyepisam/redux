# A hello world example for redux

Redux is a top down software build tool. The best way to explain it is to build something with it.
In this example, we are going to build a `hello world` program in C. We'll assume that you have read the
documentation for redux and know some C, but the example should be useful in any case. If you are reading
this from a source directory, you should have all the accompanying files and can play with them. If you are
not, the source files are available [here](https://github.com/gyepisam/blob/master/doc/examples/hello-world-0)


Our basic C hello world program, which consists of a couple of headers and a function, looks like this:

.pull hello.c,code

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

.pull hello.do,code

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

To recap, in this example, we created a simple hello world project that uses redux in order to show the basic steps of process
and used redo-init, redo-ifchange and redo. The final command, redo-ifcreate, is not as common and can be left off until later.

The only other useful item to know about are default do files. However, that is covered in the redo documentation and I really
want you to read it, so I'll point you [there](https://github.com/jireva/redux/blob/master/doc/redo.md)







