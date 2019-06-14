package commands

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/wilriker/librfm"
	"github.com/wilriker/rfm"
)

// BaseOptions is the struct holding the basic parameters common to all commands
type BaseOptions struct {
	device   string
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

		b.fs.StringVar(&b.device, "device", rfm.DefaultDevice, "Use this device from the config file")
		b.fs.StringVar(&b.domain, "domain", "", "Domain of Duet Wifi")
		b.fs.Uint64Var(&b.port, "port", 80, "Port of Duet Wifi")
		b.fs.StringVar(&b.password, "password", "reprap", "Connection password")
		b.fs.BoolVar(&b.verbose, "verbose", false, "Output more details")
		b.fs.BoolVar(&b.debug, "debug", false, "Output details on underlying HTTP requests")

	})
	return b.fs
}

func (b *BaseOptions) updateFromConfig() {

	// Find out which parameters where set by the user
	var domainSeen, portSeen, passwordSeen bool

	// Using Visit is clumsy but the only way to find out
	b.GetFlagSet().Visit(func(f *flag.Flag) {
		switch f.Name {
		case "domain":
			domainSeen = true
		case "port":
			portSeen = true
		case "password":
			passwordSeen = true
		}
	})

	c, _ := rfm.GetConfigs()
	if c != nil {

		// Existing config found
		if d := c.GetDevice(b.device); d != nil {
			if !domainSeen {
				b.domain = d.Domain
			} else {
				d.Domain = b.domain
			}
			if !portSeen {
				b.port = d.Port
			} else {
				d.Port = b.port
			}
			if !passwordSeen {
				b.password = d.Password
			} else {
				d.Password = b.password
			}
		}
	}
	rfm.AddConfig(b.device, b.domain, b.port, b.password)
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
func (b *BaseOptions) Connect() {
	b.Rfm = librfm.New(b.domain, b.port, b.debug)
	if err := b.Rfm.Connect(b.password); err != nil {
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
