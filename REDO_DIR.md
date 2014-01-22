* The .redo/data directory

The .redo directory marks the project root directory and holds configuration and data files.

.redo/data is a directory

entries are directories whose names are sha1 checksums of file paths relative to the .redo directory.

Each entry could have up to four contents

METADATA is a JSON file containing the following data: 

 path
 sha1sum

REBUILD is an empty file created to denote that the target is out of date.

satisfies and requires are directories which denote the target's relationships.
