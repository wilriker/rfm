package commands

import (
	"context"
	"log"
	"os"

	"strings"

	"github.com/wilriker/rfm"
)

// DownloadOptions holds the specific parameters for downloads
type DownloadOptions struct {
	*BaseOptions
	remotePath string
	localName  string
}

// Check checks all parameters for valid values
func (d *DownloadOptions) Check() {
	d.BaseOptions.Check()

	d.remotePath = rfm.CleanRemotePath(d.remotePath)
	if d.remotePath == "" {
		log.Fatal("<remote/file> is mandatory")
	}

	// Use same name as remote file if nothing is specified here
	if d.localName == "" {
		s := strings.Split(d.remotePath, "/")
		d.localName = s[len(s)-1]
	}
	d.localName = rfm.GetAbsPath(d.localName)
}

// InitDownloadOptions initializes a DownloadOptions instance from command-line parameters
func InitDownloadOptions(ctx context.Context, arguments []string) *DownloadOptions {
	d := DownloadOptions{BaseOptions: &BaseOptions{}}

	fs := d.GetFlagSet()
	fs.Parse(arguments)

	l := fs.NArg()
	if l > 0 {
		d.remotePath = fs.Arg(0)
		if l > 1 {
			d.localName = fs.Arg(1)
		}
	}

	d.Check()

	d.Connect(ctx)

	return &d
}

// DoDownload is a convenience method to run a download form command-line parameters
func DoDownload(ctx context.Context, arguments []string) error {
	do := InitDownloadOptions(ctx, arguments)
	return NewDownload(do).Download(ctx, do.remotePath, do.localName)
}

// download implements the Download interface
type download struct {
	o *DownloadOptions
}

// NewDownload creates a new instance of the Download interface
func NewDownload(do *DownloadOptions) *download {
	return &download{
		o: do,
	}
}

// Download downloads a remote file to a local path
func (d *download) Download(ctx context.Context, remotePath, localName string) error {
	content, duration, err := d.o.Rfm.Download(ctx, remotePath)
	if err != nil {
		return err
	}

	// Create corresponding local file
	nf, err := os.Create(localName)
	if err != nil {
		return err
	}
	defer nf.Close()

	// Write contents to local file
	_, err = nf.Write(content)
	if err != nil {
		return err
	}

	if d.o.verbose {
		kibs := (float64(len(content)) / duration.Seconds()) / 1024
		log.Printf("Downloaded: %s to %s (%.1f KiB/s)", remotePath, localName, kibs)
	}

	return nil
}
