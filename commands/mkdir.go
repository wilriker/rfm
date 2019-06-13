package commands

import (
	"log"

	"github.com/wilriker/rfm"
)

//MkdirOptions hold the specific parameters for mkdir
type MkdirOptions struct {
	*BaseOptions
	path string
}

// Init inialies a MkdirOptions instance from command-line parameters
func (m *MkdirOptions) Init(arguments []string) {
	if m.BaseOptions == nil {
		m.BaseOptions = &BaseOptions{}
	}
	fs := m.GetFlagSet()
	fs.Parse(arguments)

	if len(fs.Args()) > 0 {
		m.path = fs.Arg(0)
	}

	m.Check()

	m.Connect()
}

// Check checks all parameters for valid values
func (m *MkdirOptions) Check() {
	m.BaseOptions.Check()

	if m.path == "" {
		log.Fatal("remote path is mandatory")
	}
	m.path = rfm.CleanRemotePath(m.path)
}

// DoMkdir is a convenience function to run mkdir from command-line parameters
func DoMkdir(arguments []string) error {
	mo := &MkdirOptions{}
	mo.Init(arguments)

	m := NewMkdir(mo)

	return m.Mkdir(mo.path)
}

// Mkdir provides a single method to run mkdir
type Mkdir interface {
	Mkdir(path string) error
}

// mkdir implements the Mkdir interface
type mkdir struct {
	o *MkdirOptions
}

// NewMkdir creates a new isntance of the Mkdir interface
func NewMkdir(mo *MkdirOptions) Mkdir {
	return &mkdir{
		o: mo,
	}
}

// Mkdir creates new remote directory if it does not exist yet
func (m *mkdir) Mkdir(path string) error {
	if m.o.verbose {
		log.Println("Creating directory", path)
	}
	return m.o.Rfm.Mkdir(path)
}
