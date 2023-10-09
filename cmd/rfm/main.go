package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wilriker/rfm/commands"
)

func main() {
	if len(os.Args) < 2 {
		commands.PrintHelp(commands.NoParameters, 1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()
	var err error
	switch os.Args[1] {
	case "backup":
		err = commands.DoBackup(ctx, os.Args[2:])
	case "upload":
		err = commands.DoUpload(ctx, os.Args[2:])
	case "mkdir":
		err = commands.DoMkdir(ctx, os.Args[2:])
	case "mv":
		err = commands.DoMv(ctx, os.Args[2:])
	case "rm":
		err = commands.DoRm(ctx, os.Args[2:])
	case "download":
		err = commands.DoDownload(ctx, os.Args[2:])
	case "fileinfo":
		err = commands.DoFileinfo(ctx, os.Args[2:])
	case "ls":
		err = commands.DoLs(ctx, os.Args[2:])
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
