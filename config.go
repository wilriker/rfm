package rfm

import (
	"path/filepath"

	"os"

	"io/ioutil"

	"sync"

	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml"
)

const (
	// ConfigFileName is the name of config file
	ConfigFileName = "rfm.toml"
	// DefaultDevice is the name of the default device
	DefaultDevice = "default"
	// DefaultPort is to be used if the user did not pass a port
	DefaultPort = 80
	// DefaultPassword is used if the user did not pass a password
	DefaultPassword = "reprap"
)

type device struct {
	Domain   string
	Port     uint64
	Password string
	Excludes map[string]Excludes
}

// Config holds the configuration sets
type Config struct {
	// Devices is only exported for marshalling/unmarshalling. Use GetDevice(string) instead
	Devices map[string]device
}

var conf = &Config{}
var mu sync.Mutex
var load sync.Once

// GetDevice returns a pointer to the config for the given device name.
// Even though Config.Devices is exported this is the preferred way
// to fetch a device
func GetDevice(deviceName string) *device {
	loadConfigs()
	d, ok := conf.Devices[deviceName]
	if !ok {
		return nil
	}
	if d.Excludes == nil {
		d.Excludes = make(map[string]Excludes)
	}
	return &d
}

// LoadConfigs tries to read the config file and returns
// either its contents or an empty config and in case of
// an error also an error instance
func loadConfigs() (*Config, error) {
	var err error
	load.Do(func() {
		mu.Lock()
		defer mu.Unlock()

		// Get the user's home dir
		h, err := homedir.Dir()
		if err != nil {
			return
		}

		// Try to open the file
		f, err := os.Open(filepath.Join(h, ConfigFileName))
		if err != nil {

			// If it does not exist return an empty config
			if os.IsNotExist(err) {
				err = nil
				return
			}

			return
		}
		defer f.Close()

		// Read the file and unmarshal it
		err = toml.NewDecoder(f).Decode(conf)
	})
	if conf.Devices == nil {
		conf.Devices = make(map[string]device)
	}
	return conf, err
}

// AddConfig adds a new device to the configuration
func AddConfig(deviceName string, domain string, port uint64, password string) {
	loadConfigs()
	mu.Lock()
	defer mu.Unlock()
	d := device{
		Domain:   domain,
		Port:     port,
		Password: password,
		Excludes: make(map[string]Excludes),
	}
	conf.Devices[deviceName] = d
}

// SaveConfigs writes all known configurations to the config file
func SaveConfigs() error {
	mu.Lock()
	defer mu.Unlock()

	// Marshal the config
	bytes, err := toml.Marshal(conf)
	if err != nil {
		return err
	}

	// Get the user's home directory
	h, err := homedir.Dir()
	if err != nil {
		return err
	}

	// Create a temporary file to not kill current contents in case of error
	f, err := ioutil.TempFile(h, ConfigFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	// And write to it
	if _, err = f.Write(bytes); err != nil {
		return err
	}

	// Explicitely close the file before renaming
	f.Close()

	// If we get here rename the temporary file to the real name
	return os.Rename(f.Name(), filepath.Join(h, ConfigFileName))
}
