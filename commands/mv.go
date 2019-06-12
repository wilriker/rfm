package commands

import (
	"log"

	"github.com/wilriker/rfm"
)

// MvOptions holds the specific parameters for mv
type MvOptions struct {
	*BaseOptions
	oldpath      string
	newpath      string
	removeTarget bool
}

// Init initializes a new MvOptions instance from command-line parameters
func (m *MvOptions) Init(arguments []string) {
	if m.BaseOptions == nil {
		m.BaseOptions = &BaseOptions{}
	}
	fs := m.GetFlagSet()
	fs.StringVar(&m.oldpath, "oldpath", "", "Current name of the file/directory to rename/move")
	fs.StringVar(&m.newpath, "newpath", "", "New name/location of the file/diretory to rename/move")
	fs.BoolVar(&m.removeTarget, "overwrite", false, "Overwrite the file with <newname>")
	fs.Parse(arguments)

	m.Check()

	m.Connect()
}

// Check checks ll parameters for valid values
func (m *MvOptions) Check() {
	m.BaseOptions.Check()

	m.oldpath = rfm.CleanRemotePath(m.oldpath)
	m.newpath = rfm.CleanRemotePath(m.newpath)
	if m.oldpath == "" || m.newpath == "" {
		log.Fatal("-oldpath and -newpath are mandatory")
	}
}

// DoMv is a convenience function to run mv from command-line parameters
func DoMv(arguments []string) error {
	mo := &MvOptions{}
	mo.Init(arguments)

	m := NewMv(mo)

	return m.Mv(mo.oldpath, mo.newpath, mo.removeTarget)
}

// Mv provides a single method to move/rename a file/directory
type Mv interface {
	Mv(oldpath string, newpath string, overwrite bool) error
}

// mv implements the Mv interface
type mv struct {
	o *MvOptions
}

// NewMv creates a new instance of the Mv interface
func NewMv(mo *MvOptions) Mv {
	return &mv{
		o: mo,
	}
}

// Mv renames or moves a file or directory within a drive
func (m *mv) Mv(oldpath, newpath string, removeTarget bool) error {
	if !removeTarget {
		return m.o.Rfm.Move(oldpath, newpath)
	}
	if m.o.verbose {
		log.Println("Checking existence of", newpath)
	}
	if _, err := m.o.Rfm.Fileinfo(newpath); err == nil {
		if m.o.verbose {
			log.Println("Deleting", newpath)
		}
		if err := m.o.Rfm.Delete(newpath); err != nil {
			return err
		}
	}
	return m.o.Rfm.Move(oldpath, newpath)
}
