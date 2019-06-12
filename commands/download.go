package commands

import (
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

// Init initializes a DownloadOptions instance from command-line parameters
func (d *DownloadOptions) Init(arguments []string) {
	if d.BaseOptions == nil {
		d.BaseOptions = &BaseOptions{}
	}
	fs := d.GetFlagSet()
	fs.StringVar(&d.remotePath, "remotePath", "", "Remote path of file to download")
	fs.StringVar(&d.localName, "localName", "", "Local name of file")
	fs.Parse(arguments)

	d.Check()

	d.Connect()
}

// Check checks all parameters for valid values
func (d *DownloadOptions) Check() {
	d.BaseOptions.Check()

	d.remotePath = rfm.CleanRemotePath(d.remotePath)
	if d.remotePath == "" {
		log.Fatal("-remotePath is mandatory")
	}

	// Use same name as remote file if nothing is specified here
	if d.localName == "" {
		s := strings.Split(d.remotePath, "/")
		d.localName = s[len(s)-1]
	}
	d.localName = rfm.GetAbsPath(d.localName)
}

// DoDownload is a convenience method to run a download form command-line parameters
func DoDownload(arguments []string) error {
	do := &DownloadOptions{}
	do.Init(arguments)

	d := NewDownload(do)

	return d.Download(do.remotePath, do.localName)
}

// Download provides a single method to run a download
type Download interface {
	Download(remotePath, localName string) error
}

// download implements the Download interface
type download struct {
	o *DownloadOptions
}

// NewDownload creates a new instance of the Download interface
func NewDownload(do *DownloadOptions) Download {
	return &download{
		o: do,
	}
}

// Download downloads a remote file to a local path
func (d *download) Download(remotePath, localName string) error {
	content, duration, err := d.o.Rfm.Download(remotePath)
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
