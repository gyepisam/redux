## The redo-init command

The redo-init command creates and initializes a .redo directory,
which holds the redo configuration and database files
and also demarcates the root of the redo enabled project.

# Usage
        
        redo-init DIRECTORY

        env REDO_DIR=DIRECTORY redo-init

        redo-init

In case 1, the target directory is specified as an argument.
In case 2, it is provided by the REDO_DIR environment value.
In case 3, it not provided at all, so the current directory is used.

# Notes

The redo-init command must be invoked before any other redo commands can be used
in the project.

The command is idempotent and can be safely invoked multiple times in the same directory.

