# rfm
Command-line interface to perform file management on RepRapFirmware based devices.

## Usage
```
$ ./rfm help
rfm provides a command-line interface to perform file actions
against the HTTP interface of a device running RepRapFirmware.

Usage:
        rfm <command> [arguments]

Each command will at least expect the argument -domain which
specifies where on the network the device is located. This can
either be a resolvable hostname or an IPv4 address.

Common options to all commands:
        -domain <domain|IP>     Network address of device. Mandatory parameter.
        -port <port>            Port the device is reachable on (default 80)
        -password <password>    Connection password (default "reprap")
        -verbose                Output more details
        -debug                  Output details on underlying HTTP requests

The commands are:
        backup       Backup a given directory on the device
        upload       Upload local files/directories to the device
        mkdir        Create a new directory on the device
        mv           Rename/move a file/directory on the device
        rm           Remove a file/directory on the device
        download     Download a single file from the device
        fileinfo     Get information on a file
        ls           Show the file tree of a given path

Use "rfm help <command>" for more information about a command.
```
## Feedback
Please provide any feedback either here in the Issues or send a pull request or go to [the Duet3D forum](https://forum.duet3d.com/topic/10880/rfm-reprapfirmware-filemanager-duetbackup-successor).
