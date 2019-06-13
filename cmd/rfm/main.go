package main

import (
	"fmt"
	"log"
	"os"

	"github.com/wilriker/rfm/commands"
)

func printUsage() {
	fmt.Println(
		`rfm provides a command-line interface to perform file actions
against the HTTP interface of a device running RepRapFirmware.

Usage:
        rfm <command> [arguments]

Each command will at least expect the argument -domain which
specifies where on the network the device is located. This can
either be a resolvable hostname or an IPv4 address.

The commands are:
        backup     Backup a given directory on the device
        upload     Upload local files/directories to the device
        mkdir      Create a new directory on the device
        mv         Rename/move a file/directory on the device
        rm         Remove a file/directory on the device
        download   Download a single file from the device
        fileinfo   Get information on a file
        ls         Show the file tree of a given path

Use "rfm <command> -help" for more information about a command's
arguments.`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	var err error
	switch os.Args[1] {
	case "backup":
		err = commands.DoBackup(os.Args[2:])
	case "upload":
		err = commands.DoUpload(os.Args[2:])
	case "mkdir":
		err = commands.DoMkdir(os.Args[2:])
	case "mv":
		err = commands.DoMv(os.Args[2:])
	case "rm":
		err = commands.DoRm(os.Args[2:])
	case "download":
		err = commands.DoDownload(os.Args[2:])
	case "fileinfo":
		err = commands.DoFileinfo(os.Args[2:])
	case "ls":
		err = commands.DoLs(os.Args[2:])
	case "help":
		printUsage()
	default:
		err = fmt.Errorf("Unknown command: %s", os.Args[1])
	}
	if err != nil {
		log.Fatal(err)
	}
}
