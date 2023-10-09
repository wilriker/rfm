package commands

import (
	"context"
	"flag"
	"log"
	"os"
	"sync"

	"github.com/wilriker/librfm/v2"
	"github.com/wilriker/rfm"
)

// BaseOptions is the struct holding the basic parameters common to all commands
type BaseOptions struct {
	device      string
	domain      string
	port        uint64
	password    string
	verbose     bool
	debug       bool
	optionsSeen map[string]bool
	fs          *flag.FlagSet
	once        sync.Once
	Rfm         *librfm.RRFFileManager
}

// GetFlagSet returns the basic flag.FlagSet shared by all commands
func (b *BaseOptions) GetFlagSet() *flag.FlagSet {
	b.once.Do(func() {
		b.fs = flag.NewFlagSet("options", flag.ExitOnError)

		b.fs.StringVar(&b.device, "device", rfm.DefaultDevice, "Use this device from the config file")
		b.fs.StringVar(&b.domain, "domain", "", "Domain of Duet Wifi")
		b.fs.Uint64Var(&b.port, "port", 80, "Port of Duet Wifi")
		b.fs.StringVar(&b.password, "password", "reprap", "Connection password")
		b.fs.BoolVar(&b.verbose, "verbose", false, "Output more details")
		b.fs.BoolVar(&b.debug, "debug", false, "Output details on underlying HTTP requests")

	})
	return b.fs
}

func (b *BaseOptions) initOptionsSeen() {

	b.optionsSeen = make(map[string]bool)

	// Using Visit is clumsy but the only way to find out
	b.GetFlagSet().Visit(func(f *flag.Flag) {
		b.optionsSeen[f.Name] = true
	})
}

func (b *BaseOptions) updateFromConfig() {
	b.initOptionsSeen()

	// Get possibly existing config
	if d := rfm.GetDevice(b.device); d != nil {
		if !b.optionsSeen["domain"] {
			b.domain = d.Domain
		} else {
			d.Domain = b.domain
		}
		if !b.optionsSeen["port"] {
			b.port = d.Port
		} else {
			d.Port = b.port
		}
		if !b.optionsSeen["password"] {
			b.password = d.Password
		} else {
			d.Password = b.password
		}
	} else {
		rfm.AddConfig(b.device, b.domain, b.port, b.password)
	}
}

// Check checks the basic parameters for correctness
func (b *BaseOptions) Check() {

	// Check port first
	if b.port > 65535 {
		log.Fatal("Invalid port: ", b.port)
	}

	// Update settings from config and config from parameters
	b.updateFromConfig()
	if b.domain == "" {
		log.Fatal("-domain is mandatory")
	}

}

// Connect initializes the connection to RepRapFirmware
func (b *BaseOptions) Connect(ctx context.Context) {
	b.Rfm = librfm.New(b.domain, b.port, b.debug)
	if err := b.Rfm.Connect(ctx, b.password); err != nil {
		log.Println("Duet currently not available")
		os.Exit(0)
	}
	// Save config after successful connect
	err := rfm.SaveConfigs()
	// Inform user about problem saving file but don't stop
	if err != nil {
		log.Printf("Unable to save configuration for %s to %s: %s", b.device, rfm.ConfigFileName, err)
	}
}
