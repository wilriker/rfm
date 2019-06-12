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

// Init intitializes a new UploadOptions instance from command-line parameters
func (u *UploadOptions) Init(arguments []string) {
	if u.BaseOptions == nil {
		u.BaseOptions = &BaseOptions{}
	}
	fs := u.GetFlagSet()
	fs.StringVar(&u.localPath, "localPath", "", "Local path to file or directory being uploaded")
	fs.StringVar(&u.remotePath, "remotePath", "", "Remote name of file or directory being uploaded")
	fs.Var(&u.excls, "exclude", "Exclude paths starting with this string (can be passed multiple times)")
	fs.Parse(arguments)

	u.Check()

	u.Connect()
}

// Check checks all parameters for valid values
func (u *UploadOptions) Check() {
	u.BaseOptions.Check()

	u.localPath = rfm.GetAbsPath(u.localPath)
	u.remotePath = rfm.CleanRemotePath(u.remotePath)
	if u.localPath == "" || u.remotePath == "" {
		log.Fatal("-localPath and -remotePath are mandatory")
	}
	u.excls.ForEach(rfm.GetAbsPath)
}

// DoUpload is a convencience function to run upload from command-line parameters
func DoUpload(arguments []string) error {
	uo := &UploadOptions{}
	uo.Init(arguments)

	u := NewUpload(uo)

	return u.Upload(uo.localPath, uo.remotePath)
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
	var err error
	err = filepath.Walk(u.o.localPath, func(path string, info os.FileInfo, err error) error {
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

		lp := strings.TrimPrefix(path, u.o.localPath)
		rp := rfm.CleanRemotePath(fmt.Sprintf("%s/%s", remotePath, lp))

		// Directories are created automatically where necessary
		if info.IsDir() {
			return nil
		}

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
	return err
}
