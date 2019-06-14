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

// Check checks ll parameters for valid values
func (m *MvOptions) Check() {
	m.BaseOptions.Check()

	if m.oldpath == "" || m.newpath == "" {
		log.Fatal("<old/path> and <new/path> are mandatory")
	}
	m.oldpath = rfm.CleanRemotePath(m.oldpath)
	m.newpath = rfm.CleanRemotePath(m.newpath)
}

// InitMvOptions initializes a new MvOptions instance from command-line parameters
func InitMvOptions(arguments []string) *MvOptions {
	m := MvOptions{BaseOptions: &BaseOptions{}}

	fs := m.GetFlagSet()
	fs.BoolVar(&m.removeTarget, "f", false, "Overwrite the file with <newname>")
	fs.Parse(arguments)

	l := fs.NArg()
	if l > 0 {
		m.oldpath = fs.Arg(0)
		if l > 1 {
			m.newpath = fs.Arg(1)
		}
	}

	m.Check()

	m.Connect()

	return &m
}

// DoMv is a convenience function to run mv from command-line parameters
func DoMv(arguments []string) error {
	mo := InitMvOptions(arguments)
	return NewMv(mo).Mv(mo.oldpath, mo.newpath, mo.removeTarget)
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
