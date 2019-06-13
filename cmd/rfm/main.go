package main

import (
	"fmt"
	"log"
	"os"

	"github.com/wilriker/rfm/commands"
)

func main() {
	if len(os.Args) < 2 {
		commands.PrintHelp(commands.NoParameters, 1)
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
		if len(os.Args) > 2 {
			commands.PrintHelp(os.Args[2:], 0)
		}
		commands.PrintHelp(commands.NoParameters, 0)
	default:
		err = fmt.Errorf("Unknown command: %s", os.Args[1])
	}
	if err != nil {
		log.Fatal(err)
	}
}
