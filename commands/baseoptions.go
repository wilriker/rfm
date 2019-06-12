package commands

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/wilriker/librfm"
)

// BaseOptions is the struct holding the basic parameters common to all commands
type BaseOptions struct {
	domain   string
	port     uint64
	password string
	verbose  bool
	debug    bool
	fs       *flag.FlagSet
	once     sync.Once
	Rfm      librfm.RRFFileManager
}

// GetFlagSet returns the basic flag.FlagSet shared by all commands
func (b *BaseOptions) GetFlagSet() *flag.FlagSet {
	b.once.Do(func() {
		b.fs = flag.NewFlagSet("options", flag.ExitOnError)

		b.fs.StringVar(&b.domain, "domain", "", "Domain of Duet Wifi")
		b.fs.Uint64Var(&b.port, "port", 80, "Port of Duet Wifi")
		b.fs.StringVar(&b.password, "password", "reprap", "Connection password")
		b.fs.BoolVar(&b.verbose, "verbose", false, "Output more details")
		b.fs.BoolVar(&b.debug, "debug", false, "Output details on underlying HTTP requests")

	})
	return b.fs
}

// Check checks the basic parameters for correctness
func (b *BaseOptions) Check() {
	if b.domain == "" {
		log.Fatal("-domain is mandatory")
	}

	if b.port > 65535 {
		log.Fatal("Invalid port: ", b.port)
	}
}

// Connect initializes the connection to RepRapFirmware
func (b *BaseOptions) Connect() {
	b.Rfm = librfm.New(b.domain, b.port, b.debug)
	if err := b.Rfm.Connect(b.password); err != nil {
		log.Println("Duet currently not available")
		os.Exit(0)
	}
}
