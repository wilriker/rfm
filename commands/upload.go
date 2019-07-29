package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wilriker/rfm"
)

// UploadOptions hold the specific parameters for upload
type UploadOptions struct {
	*BaseOptions
	localPath  string
	remotePath string
	excls      rfm.Excludes
}

// Check checks all parameters for valid values
func (u *UploadOptions) Check() {
	u.BaseOptions.Check()

	u.localPath = rfm.GetAbsPath(u.localPath)
	u.remotePath = rfm.CleanRemotePath(u.remotePath)

	d := rfm.GetDevice(u.device)
	if !u.optionsSeen["exclude"] {
		u.excls = d.Excludes["upload"]
	} else {
		d.Excludes["upload"] = u.excls
	}
	u.excls.ForEach(rfm.GetAbsPath)
}

// InitUploadOptions intitializes a new UploadOptions instance from command-line parameters
func InitUploadOptions(arguments []string) *UploadOptions {
	u := UploadOptions{BaseOptions: &BaseOptions{}}

	fs := u.GetFlagSet()
	fs.Var(&u.excls, "exclude", "Exclude paths starting with this string (can be passed multiple times)")
	fs.Parse(arguments)

	l := fs.NArg()
	if l > 0 {
		u.localPath = fs.Arg(0)
		if l > 1 {
			u.remotePath = fs.Arg(1)
		}
	}

	u.Check()

	u.Connect()

	return &u
}

// DoUpload is a convencience function to run upload from command-line parameters
func DoUpload(arguments []string) error {
	uo := InitUploadOptions(arguments)
	return NewUpload(uo).Upload(uo.localPath, uo.remotePath)
}

// Upload provides a single method to run an upload
type Upload interface {
	Upload(localPath, remotePath string) error
}

// upload implements the Upload interface
type upload struct {
	o *UploadOptions
}

// NewUpload creates a new instance of the Upload interface
func NewUpload(uo *UploadOptions) Upload {
	return &upload{
		o: uo,
	}
}

// Upload uploads a file or directory (structure) to the given remote path
func (u *upload) Upload(localPath, remotePath string) error {
	return filepath.Walk(u.o.localPath, func(path string, info os.FileInfo, err error) error {
		if u.o.excls.Contains(path) {
			if info.IsDir() {
				if u.o.verbose {
					log.Println("Skipping directory", path)
				}
				return filepath.SkipDir
			}
			if u.o.verbose {
				log.Println("Skipping", path)
			}
			return nil
		}

		// Directories are created automatically where necessary
		if info.IsDir() {
			return nil
		}

		lp := strings.TrimPrefix(path, u.o.localPath)
		if lp == "" {
			lp = info.Name()
		}
		rp := rfm.CleanRemotePath(fmt.Sprintf("%s/%s", remotePath, lp))

		fmt.Println("path:", path, "\nlp:", lp, "\nremotePath:", remotePath, "\nrp:", rp)

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		if u.o.verbose {
			log.Printf("Uploading %s to %s", path, rp)
		}
		_, err = u.o.Rfm.Upload(rp, f)
		return err
	})
}
