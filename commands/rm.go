package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/wilriker/librfm/v2"
	"github.com/wilriker/rfm"
)

// RmOptions holds the specific parameters for rm
type RmOptions struct {
	*BaseOptions
	path      string
	recursive bool
}

// Check checks all parameters for valid values
func (r *RmOptions) Check() {
	r.BaseOptions.Check()

	if r.path == "" {
		log.Fatal("<remote/path> is mandatory")
	}
	r.path = rfm.CleanRemotePath(r.path)
}

// InitRmOptions initializes a new RmOptions instance from command-line parameters
func InitRmOptions(ctx context.Context, arguments []string) *RmOptions {
	r := RmOptions{BaseOptions: &BaseOptions{}}

	fs := r.GetFlagSet()
	fs.BoolVar(&r.recursive, "r", false, "Remove recursively")
	fs.Parse(arguments)

	if fs.NArg() > 0 {
		r.path = fs.Arg(0)
	}

	r.Check()

	r.Connect(ctx)

	return &r
}

// DoRm is a convenience function to run rm from command-line parameters
func DoRm(ctx context.Context, arguments []string) error {
	ro := InitRmOptions(ctx, arguments)
	return NewRm(ro).Rm(ctx, ro.path, ro.recursive)
}

// rm impelements the Rm interface
type rm struct {
	o *RmOptions
}

// NewRm creates a new instance of the Rm interface
func NewRm(r *RmOptions) *rm {
	return &rm{
		o: r,
	}
}

// Rm deletes a file or directory.
// Directorries will only be removed if empty or together
// with all their contents if recursive is true.
func (r *rm) Rm(ctx context.Context, path string, recursive bool) error {
	if !recursive {
		if r.o.verbose {
			log.Println("Deleting", path)
		}
		return r.o.Rfm.Delete(ctx, path)
	}
	fl, err := r.o.Rfm.Filelist(ctx, path, true)
	if err != nil {
		return err
	}
	if err = r.deleteRecursive(ctx, fl); err != nil {
		return err
	}
	if r.o.verbose {
		log.Println("Deleting", fl.Dir)
	}
	return r.o.Rfm.Delete(ctx, fl.Dir)
}

func (r *rm) deleteRecursive(ctx context.Context, fl *librfm.Filelist) error {
	for _, f := range fl.Subdirs {
		if err := r.deleteRecursive(ctx, f); err != nil {
			return err
		}
	}
	for _, f := range fl.Files {
		remotePath := fmt.Sprintf("%s/%s", fl.Dir, f.Name)
		if r.o.verbose {
			log.Println("Deleting", remotePath)
		}
		if err := r.o.Rfm.Delete(ctx, remotePath); err != nil {
			return err
		}
	}
	return nil
}
