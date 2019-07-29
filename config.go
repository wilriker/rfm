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

// Config holds the configuration sets
type Config struct {
	// Devices is only exported for marshalling/unmarshalling. Use GetDevice(string) instead
	Devices map[string]device
}

func (c *Config) init() {
	if c.Devices == nil {
		c.Devices = make(map[string]device)
	}
}

// GetDevice returns a pointer to the config for the given device name.
// Even though Config.Devices is exported this is the preferred way
// to fetch a device
func (c *Config) GetDevice(deviceName string) *device {
	c.init()
	if d, ok := c.Devices[deviceName]; ok {
		return &d
	}
	return nil
}

type device struct {
	Domain   string
	Port     uint64
	Password string
}

var conf = &Config{}
var mu sync.Mutex

// GetConfigs tries to read the config file and returns
// either its contents or an empty config and in case of
// an error also an error instance
func GetConfigs() (*Config, error) {
	mu.Lock()
	defer mu.Unlock()

	// Make sure the internal map is initialized on return
	defer conf.init()

	// Get the user's home dir
	h, err := homedir.Dir()
	if err != nil {
		return conf, err
	}

	// Try to open the file
	f, err := os.Open(filepath.Join(h, ConfigFileName))
	if err != nil {

		// If it does not exist return an empty config
		if os.IsNotExist(err) {
			return conf, nil
		}

		return conf, err
	}
	defer f.Close()

	// Read the file and unmarshal it
	err = toml.NewDecoder(f).Decode(conf)
	return conf, nil
}

// AddConfig adds a new device to the configuration
func AddConfig(deviceName string, domain string, port uint64, password string) {
	mu.Lock()
	defer mu.Unlock()
	d := device{
		Domain:   domain,
		Port:     port,
		Password: password,
	}
	conf.init()
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
