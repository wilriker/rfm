package commands

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/wilriker/librfm/v2"
	"github.com/wilriker/rfm"
)

const (
	// SysDir is the location of the default configuration directory
	SysDir           = "0:/sys"
	managedDirMarker = ".rfmbackup"
)

// BackupOptions holds all options relevant to a backup process
type BackupOptions struct {
	*BaseOptions
	dirToBackup string
	outDir      string
	removeLocal bool
	excls       rfm.Excludes
}

// Check checks all parameters for valid values
func (b *BackupOptions) Check() {
	b.BaseOptions.Check()

	b.outDir = rfm.GetAbsPath(b.outDir)
	b.dirToBackup = rfm.CleanRemotePath(b.dirToBackup)

	d := rfm.GetDevice(b.device)
	if !b.optionsSeen["exclude"] {
		b.excls = d.Excludes["backup"]
	} else {
		d.Excludes["backup"] = b.excls
	}

	b.excls.ForEach(rfm.CleanRemotePath)
}

// InitBackupOptions intializes a backupOptions instance from command line parameters
func InitBackupOptions(ctx context.Context, arguments []string) *BackupOptions {
	b := BackupOptions{BaseOptions: &BaseOptions{}}

	fs := b.GetFlagSet()
	fs.BoolVar(&b.removeLocal, "removeLocal", false, "Remove files locally that have been deleted on the Duet")
	fs.Var(&b.excls, "exclude", "Exclude paths starting with this string (can be passed multiple times)")
	if err := fs.Parse(arguments); err != nil {
		log.Fatalf("Error parsing command-line arguments: %s", err)
	}

	b.dirToBackup = SysDir
	l := fs.NArg()
	if l > 0 {
		b.outDir = fs.Arg(0)
		if l > 1 {
			b.dirToBackup = fs.Arg(1)
		}
	}

	b.Check()

	b.Connect(ctx)

	return &b
}

// DoBackup is a convenience function to run a backup from command line parameters
func DoBackup(ctx context.Context, arguments []string) error {
	bo := InitBackupOptions(ctx, arguments)
	return NewBackup(bo).Backup(ctx, bo.dirToBackup, bo.outDir, bo.excls, bo.removeLocal)
}

// backup implementes the Backup interface
type backup struct {
	o *BackupOptions
}

// NewBackup creates a new instance of the Backup interface
func NewBackup(bo *BackupOptions) *backup {
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

func (b *backup) updateLocalFiles(ctx context.Context, fl *librfm.Filelist, outDir string, excls rfm.Excludes, removeLocal bool) error {

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
			body, duration, err := b.o.Rfm.Download(ctx, remoteFilename)
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
func (b *backup) isManagedDirectory(basePath string, f fs.DirEntry) bool {
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
	existingFiles := make(map[string]bool)
	for _, f := range fl.Files {
		existingFiles[f.Name] = true
	}

	dirEntries, err := os.ReadDir(outDir)
	if err != nil {
		return err
	}

	for _, de := range dirEntries {
		if !existingFiles[de.Name()] {

			// Skip directories not managed by us as well as our marker file
			if (de.IsDir() && !b.isManagedDirectory(outDir, de)) || de.Name() == managedDirMarker {
				continue
			}
			if err := os.RemoveAll(filepath.Join(outDir, de.Name())); err != nil {
				return err
			}
			if b.o.verbose {
				marker := fileMarker
				if de.IsDir() {
					marker = dirMarker
				}
				log.Println("  Removed:   ", marker, de.Name())
			}
		}
	}

	return nil
}

// Backup will synchronize the contents of a remote folder to a local directory.
// The boolean flag removeLocal decides whether or not files that have been remove
// remote should also be deleted locally
func (b *backup) Backup(ctx context.Context, folder, outDir string, excls rfm.Excludes, removeLocal bool) error {

	// Skip complete directories if they are covered by an exclude pattern
	if excls.Contains(folder) {
		log.Println("Excluding", folder)
		return nil
	}

	log.Println("Fetching filelist for", folder)
	fl, err := b.o.Rfm.Filelist(ctx, folder, false)
	if err != nil {
		return err
	}

	log.Println("Downloading new/changed files from", folder, "to", outDir)
	if err = b.updateLocalFiles(ctx, fl, outDir, excls, removeLocal); err != nil {
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
		if err = b.Backup(ctx, remoteFilename, fileName, excls, removeLocal); err != nil {
			return err
		}
	}

	return nil
}
