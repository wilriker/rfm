package commands

import (
	"context"
	"fmt"

	"github.com/wilriker/librfm/v2"
	"github.com/wilriker/rfm"
)

const (
	dirMarker         = "[d]"
	fileMarker        = "[f]"
	sizePlaceHolder   = "         -"
	sizePlaceHolderHR = "     -"
)

// LsOptions holds the specific parameters for ls
type LsOptions struct {
	*BaseOptions
	paths         []string
	recursive     bool
	humanReadable bool
}

// Check checks all parameters for valid values
func (l *LsOptions) Check() {
	l.BaseOptions.Check()
	if len(l.paths) == 0 {
		l.paths = append(l.paths, "")
	}
	for i := 0; i < len(l.paths); i++ {
		l.paths[i] = rfm.CleanRemotePath(l.paths[i])
	}
}

// InitLsOptions initializes a LsOptions instance from command-line parameters
func InitLsOptions(ctx context.Context, arguments []string) *LsOptions {
	l := LsOptions{BaseOptions: &BaseOptions{}}

	fs := l.GetFlagSet()
	fs.BoolVar(&l.recursive, "r", false, "List recursively")
	fs.BoolVar(&l.humanReadable, "h", false, "List sizes in human readable units")
	fs.Parse(arguments)

	l.paths = fs.Args()

	l.Check()

	l.Connect(ctx)

	return &l
}

// DoLs is a convenience function to run ls from command-line parameters
func DoLs(ctx context.Context, arguments []string) error {
	lo := InitLsOptions(ctx, arguments)
	return NewLs(lo).Ls(ctx, lo.paths, lo.recursive)
}

// ls implements the Ls interface
type ls struct {
	o *LsOptions
}

// NewLs creates new instance of the Ls interface
func NewLs(lo *LsOptions) *ls {
	return &ls{
		o: lo,
	}
}

// Ls lists all files and directories in a given remote directory,
// optionally recursive and with human-readable sizes
func (l *ls) Ls(ctx context.Context, paths []string, recursive bool) error {
	for _, path := range paths {
		if l.o.recursive || len(paths) > 1 {
			fmt.Printf("\n%s:\n", path)
		}
		fl, err := l.o.Rfm.Filelist(ctx, path, recursive)
		if err != nil {
			return err
		}

		l.print(fl)

		if l.o.recursive {
			for _, subdir := range fl.Subdirs {
				fmt.Printf("\n%s:\n", subdir.Dir)
				l.print(subdir)
			}
		}
	}

	return nil
}

func (l *ls) print(fl *librfm.Filelist) {
	totalBytes := uint64(0)
	for _, f := range fl.Files {
		totalBytes += f.Size
	}
	fmt.Println("total", l.getSize(totalBytes))
	for _, f := range fl.Files {
		fmt.Printf("%s\t%s\t%s\t%s\n", l.getMarker(f), l.getSizeForFile(f), f.Date().Format(librfm.TimeFormat), f.Name)
	}
}

func (l *ls) getMarker(f librfm.File) string {
	if f.IsDir() {
		return dirMarker
	}
	return fileMarker
}

func (l *ls) getSize(size uint64) string {
	if l.o.humanReadable {
		return rfm.HumanReadableSize(size)
	}
	return fmt.Sprintf("%10d", size)
}

func (l *ls) getSizeForFile(f librfm.File) string {
	if f.IsDir() {
		if l.o.humanReadable {
			return sizePlaceHolderHR
		}
		return sizePlaceHolder
	}
	return l.getSize(f.Size)
}
