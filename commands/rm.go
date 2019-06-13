package commands

import (
	"fmt"
	"log"

	"github.com/wilriker/librfm"
	"github.com/wilriker/rfm"
)

// RmOptions holds the specific parameters for rm
type RmOptions struct {
	*BaseOptions
	path      string
	recursive bool
}

// Init initializes a new RmOptions instance from command-line parameters
func (r *RmOptions) Init(arguments []string) {
	if r.BaseOptions == nil {
		r.BaseOptions = &BaseOptions{}
	}
	fs := r.GetFlagSet()
	fs.BoolVar(&r.recursive, "r", false, "Remove recursively")
	fs.Parse(arguments)

	if len(fs.Args()) > 0 {
		r.path = fs.Arg(0)
	}

	r.Check()

	r.Connect()
}

// Check checks all parameters for valid values
func (r *RmOptions) Check() {
	r.BaseOptions.Check()

	if r.path == "" {
		log.Fatal("<remote/path> is mandatory")
	}
	r.path = rfm.CleanRemotePath(r.path)
}

// DoRm is a convenience function to run rm from command-line parameters
func DoRm(arguments []string) error {
	ro := &RmOptions{}
	ro.Init(arguments)

	r := NewRm(ro)

	return r.Rm(ro.path, ro.recursive)
}

// Rm provides a single method ro remove a file or directory
type Rm interface {
	Rm(path string, recursive bool) error
}

// rm impelements the Rm interface
type rm struct {
	o *RmOptions
}

// NewRm creates a new instance of the Rm interface
func NewRm(r *RmOptions) Rm {
	return &rm{
		o: r,
	}
}

// Rm deletes a file or directory.
// Directorries will only be removed if empty or together
// with all their contents if recursive is true.
func (r *rm) Rm(path string, recursive bool) error {
	if !recursive {
		if r.o.verbose {
			log.Println("Deleting", path)
		}
		return r.o.Rfm.Delete(path)
	}
	fl, err := r.o.Rfm.Filelist(path, true)
	if err != nil {
		return err
	}
	if err = r.deleteRecursive(fl); err != nil {
		return err
	}
	if r.o.verbose {
		log.Println("Deleting", fl.Dir)
	}
	return r.o.Rfm.Delete(fl.Dir)
}

func (r *rm) deleteRecursive(fl *librfm.Filelist) error {
	for _, f := range fl.Subdirs {
		if err := r.deleteRecursive(f); err != nil {
			return err
		}
	}
	for _, f := range fl.Files {
		remotePath := fmt.Sprintf("%s/%s", fl.Dir, f.Name)
		if r.o.verbose {
			log.Println("Deleting", remotePath)
		}
		if err := r.o.Rfm.Delete(remotePath); err != nil {
			return err
		}
	}
	return nil
}
