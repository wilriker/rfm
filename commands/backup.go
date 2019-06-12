package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wilriker/librfm"
	"github.com/wilriker/rfm"
)

const (
	// SysDir is the location of the default configuration directory
	SysDir           = "0:/sys"
	managedDirMarker = ".duetbackup"
)

// BackupOptions holds all options relevant to a backup process
type BackupOptions struct {
	*BaseOptions
	dirToBackup string
	outDir      string
	removeLocal bool
	excls       rfm.Excludes
}

// Init intializes a backupOptions instance from command line parameters
func (b *BackupOptions) Init(arguments []string) {
	if b.BaseOptions == nil {
		b.BaseOptions = &BaseOptions{}
	}
	fs := b.GetFlagSet()
	fs.StringVar(&b.dirToBackup, "dirToBackup", SysDir, "Directory on Duet to create a backup of")
	fs.StringVar(&b.outDir, "outDir", "", "Output dir of backup")
	fs.BoolVar(&b.removeLocal, "removeLocal", false, "Remove files locally that have been deleted on the Duet")
	fs.Var(&b.excls, "exclude", "Exclude paths starting with this string (can be passed multiple times)")
	fs.Parse(arguments)

	b.Check()

	b.Connect()
}

// Check checks all parameters for valid values
func (b *BackupOptions) Check() {
	b.BaseOptions.Check()

	b.outDir = rfm.GetAbsPath(b.outDir)
	b.dirToBackup = rfm.CleanRemotePath(b.dirToBackup)
	if b.outDir == "" {
		log.Fatal("-outDir is mandatory")
	}
	if b.dirToBackup == "" {
		log.Fatal("-dirToBackup must not be empty")
	}
	b.excls.ForEach(rfm.CleanRemotePath)
}

// DoBackup is a convenience function to run a backup from command line parameters
func DoBackup(arguments []string) error {

	bo := &BackupOptions{}
	bo.Init(arguments)

	b := NewBackup(bo)

	return b.SyncFolder(bo.dirToBackup, bo.outDir, bo.excls, bo.removeLocal)
}

// Backup provides a single method to run backups
type Backup interface {
	// SyncFolder will syncrhonize the contents of a remote folder to a local directory.
	// The boolean flag removeLocal decides whether or not files that have been remove
	// remote should also be deleted locally
	SyncFolder(remoteFolder, outDir string, excls rfm.Excludes, removeLocal bool) error
}

// backup implementes the Backup interface
type backup struct {
	o *BackupOptions
}

// NewBackup creates a new instance of the Backup interface
func NewBackup(bo *BackupOptions) Backup {
	return &backup{
		o: bo,
	}
}

// ensureOutDirExists will create the local directory if it does not exist
// and will in any case create the marker file inside it
func (b *backup) ensureOutDirExists(outDir string) error {
	path, err := filepath.Abs(outDir)
	if err != nil {
		return err
	}

	// Check if the directory exists
	fi, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Create the directory
	if fi == nil {
		if b.o.verbose {
			log.Println("  Creating directory", path)
		}
		if err = os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	// Create the marker file
	markerFile, err := os.Create(filepath.Join(path, managedDirMarker))
	if err != nil {
		return err
	}
	markerFile.Close()

	return nil
}

func (b *backup) updateLocalFiles(fl *librfm.Filelist, outDir string, excls rfm.Excludes, removeLocal bool) error {

	if err := b.ensureOutDirExists(outDir); err != nil {
		return err
	}

	for _, file := range fl.Files {
		if file.IsDir() {
			continue
		}
		remoteFilename := fmt.Sprintf("%s/%s", fl.Dir, file.Name)

		// Skip files covered by an exclude pattern
		if excls.Contains(remoteFilename) {
			if b.o.verbose {
				log.Println("  Excluding: ", remoteFilename)
			}
			continue
		}

		fileName := filepath.Join(outDir, file.Name)
		fi, err := os.Stat(fileName)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		// File does not exist or is outdated so get it
		if fi == nil || fi.ModTime().Before(file.Date()) {

			// Download file
			body, duration, err := b.o.Rfm.Download(remoteFilename)
			if err != nil {
				return err
			}

			// Create corresponding local file
			nf, err := os.Create(fileName)
			if err != nil {
				return err
			}
			defer nf.Close()

			// Write contents to local file
			_, err = nf.Write(body)
			if err != nil {
				return err
			}

			// Adjust atime and mtime
			if err = os.Chtimes(fileName, file.Date(), file.Date()); err != nil {
				return err
			}

			if b.o.verbose {
				kibs := (float64(file.Size) / duration.Seconds()) / 1024
				if fi != nil {
					log.Printf("  Updated:   %s (%.1f KiB/s)", remoteFilename, kibs)
				} else {
					log.Printf("  Added:     %s (%.1f KiB/s)", remoteFilename, kibs)
				}
			}
		} else {
			if b.o.verbose {
				log.Println("  Up-to-date:", remoteFilename)
			}
		}

	}

	return nil
}

// isManagedDirectory checks wether the given path is a directory and
// if so if it contains the marker file. It will return false in case
// any error has occured.
func (b *backup) isManagedDirectory(basePath string, f os.FileInfo) bool {
	if !f.IsDir() {
		return false
	}
	markerFile := filepath.Join(basePath, f.Name(), managedDirMarker)
	fi, err := os.Stat(markerFile)
	if err != nil && !os.IsNotExist(err) {
		return false
	}
	if fi == nil {
		return false
	}
	return true
}

func (b *backup) removeDeletedFiles(fl *librfm.Filelist, outDir string) error {

	// Pseudo hash-set of known remote filenames
	existingFiles := make(map[string]struct{})
	for _, f := range fl.Files {
		existingFiles[f.Name] = struct{}{}
	}

	files, err := ioutil.ReadDir(outDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if _, exists := existingFiles[f.Name()]; !exists {

			// Skip directories not managed by us as well as our marker file
			if !b.isManagedDirectory(outDir, f) || f.Name() == managedDirMarker {
				continue
			}
			if err := os.RemoveAll(filepath.Join(outDir, f.Name())); err != nil {
				return err
			}
			if b.o.verbose {
				log.Println("  Removed:   ", f.Name())
			}
		}
	}

	return nil
}

func (b *backup) SyncFolder(folder, outDir string, excls rfm.Excludes, removeLocal bool) error {

	// Skip complete directories if they are covered by an exclude pattern
	if excls.Contains(folder) {
		log.Println("Excluding", folder)
		return nil
	}

	log.Println("Fetching filelist for", folder)
	fl, err := b.o.Rfm.Filelist(folder, false)
	if err != nil {
		return err
	}

	log.Println("Downloading new/changed files from", folder, "to", outDir)
	if err = b.updateLocalFiles(fl, outDir, excls, removeLocal); err != nil {
		return err
	}

	if removeLocal {
		log.Println("Removing no longer existing files in", outDir)
		if err = b.removeDeletedFiles(fl, outDir); err != nil {
			return err
		}
	}

	// Traverse into subdirectories
	for _, file := range fl.Files {
		if !file.IsDir() {
			continue
		}
		remoteFilename := fmt.Sprintf("%s/%s", fl.Dir, file.Name)
		fileName := filepath.Join(outDir, file.Name)
		if err = b.SyncFolder(remoteFilename, fileName, excls, removeLocal); err != nil {
			return err
		}
	}

	return nil
}
