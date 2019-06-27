package commands

import (
	"fmt"
	"os"
)

const (
	mainHelp = `rfm provides a command-line interface to perform file actions
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
        -device <devicename>    This can be used to either create, update (both
                                in combination with the above options) or load
                                an already configured device. This makes multi-
                                device environments easier to handle.
                                (default "default")
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

Use "rfm help <command>" for more information about a command.`
	backupHelp = `Usage: rfm backup <common-options> [-removeLocal] [-exclude <excludepattern>]*
                  [<local/path> [<remote/path>]]

backup will download a directory structure from the device to a local directory.
Each locally created directory will contain a marker file named .rfmbackup.
This is important for the flag -removeLocal (see below). Directories not having
this marker file will not be removed in any case.

Options:
        -removeLocal                 Remove files locally that have been
                                     removed remote
        -exclude <excludepattern>    Exclude paths starting with this string
                                     (can be used multiple times)

Parameters:
        <local/path>     Path where the download is stored locally. If omitted
                         the current diretory is used.
        <remote/path>    Remote path to be backuped. If this is changed the
                         local path has to be provided also. (default "0:/sys")`
	uploadHelp = `Usage: rfm upload <common-options> [-exclude <excludepattern>]*
                  [<local/path> [<remote/path>]]

upload will upload a file or directory to the remote device.

Options:
        -exclude <excludepattern>    Exclude paths starting with this string
                                     (can be used multiple times)

Parameters:
        <local/path>     Local path of the file or directory to be uploaded
                         (default: current directory)
        <remote/path>    Remote path to store the file(s)/directory at
                         (default: 0:/)`
	mkdirHelp = `Usage: rfm mkdir <common-options> <remote/path>

mkdir will create a new directory on the device.

Parameters:
        <remote/path>    Remote path to be created

Errors:
This will return an error in both cases where the directory could not be created
or the directory already exists. Since both cases return the same error they
cannot be differtiated by rfm.`
	mvHelp = `Usage: rfm mv <common-options> [-f] <old/path> <new/path>

mv will move or rename a file or directory withing one mounted volume.

Options:
        -f    Overwrite the target file if it exists.
              This will not delete existing directories.

Parameters:
        <old/path>    Current path of the file or directory to be
                      moved or renamed
        <new/path>    New path of the file or directory

Errors:
Trying to move files or directories across volumes will return an error.
Another source of error might be trying to rename a file to a name of an
existing directory.`
	rmHelp = `Usage: rfm rm <common-options> [-r] <remote/path>

rm will delete a remote file or directory. Directories can only be deleted if
they are empty or the option "-r" is given which enables recursive delete.

Options:
        -r    Delete directories recursively, i.e. including ALL their contents

Parameters:
        <remote/path>    Path of the remote file or directory`
	downloadHelp = `Usage: rfm download <common-options> <remote/file> [<local/name>]

Download downloads a single file from the device to a local directory. The name
of the local file can be given instead of using the remote filename.

Parameters:
        <remote/file>    Path to the remote file
        <local/name>     Local path and filename
                         (default: current directory + remote filename)
Errors:
If the remote path is a directory or the file does not exist there will
be an error. For directories use "rfm backup" instead.`
	fileinfoHelp = `Usage: rfm fileinfo <common-options> [-h] <remote/file>

fileinfo will display information about a remote file. It will at least return
the name, size and last modification date. For GCode files it will also output
further information as far as they could be extracted.

Options:
        -h    List file sizes in human-readble units instead of byte sizes

Parameters:
        <remote/file>    Path of the file

Errors:
If the given path is a directory or the file does not exist there will
be an error.`
	lsHelp = `Usage: rfm ls <common-options> [-h] [-r] [<remote/dir>]*

ls will list the contents of a remote directory.

Options:
        -h    List file sizes in human-readble units instead of byte size.
        -r    List directories recursively starting at the given directory

Parameters:
		<remote/dir>    Remote directory to be listed. Can be used multiple
                        times. (default: 0:/)

Errors:
This will return an error in case a remote file is given as <remote/dir>
or for the first path that is not found remote.`
	unknownTopic = `rfm help %s: unknown help topic. Run 'rfm help'`
)

// NoParameters can be passed to PrintHelp if there are no further parameters
var NoParameters []string

// PrintHelp prints the help text for the appropriate command
// or outputs an error message in case an unknown help topic
// was requested
func PrintHelp(arguments []string, exitCode int) {
	if len(arguments) == 0 {
		fmt.Println(mainHelp)
		os.Exit(exitCode)
	}
	switch arguments[0] {
	case "backup":
		fmt.Println(backupHelp)
	case "upload":
		fmt.Println(uploadHelp)
	case "mkdir":
		fmt.Println(mkdirHelp)
	case "mv":
		fmt.Println(mvHelp)
	case "rm":
		fmt.Println(rmHelp)
	case "download":
		fmt.Println(downloadHelp)
	case "fileinfo":
		fmt.Println(fileinfoHelp)
	case "ls":
		fmt.Println(lsHelp)
	default:
		fmt.Printf(unknownTopic, arguments[0])
		os.Exit(1)
	}
	os.Exit(exitCode)
}
