# rfm
Command-line interface to perform file management on RepRapFirmware based devices.

## Usage
```
$ ./rfm help
rfm provides a command-line interface to perform file actions
against the HTTP interface of a device running RepRapFirmware.

Usage:
        rfm <command> [arguments]

Each command will at least expect the argument -domain which
specifies where on the network the device is located. This can
either be a resolvable hostname or an IPv4 address.

Common options to all commands:
        -domain <domain|IP>     Network address of device. Mandatory parameter.
        -port <port>            Port the device is reachable on (default 80)
        -password <password>    Connection password (default "reprap")
        -device <devicename>    This can be used to either create, update (both
                                in combination with the above options) or load
                                an already configured device. This makes multi-
                                device environments easier to handle.
                                (default "default")
        -verbose                Output more details
        -debug                  Output details on underlying HTTP requests

The commands are:
        backup       Backup a given directory on the device
        upload       Upload local files/directories to the device
        mkdir        Create a new directory on the device
        mv           Rename/move a file/directory on the device
        rm           Remove a file/directory on the device
        download     Download a single file from the device
        fileinfo     Get information on a file
        ls           Show the file tree of a given path

Use "rfm help <command>" for more information about a command.
```

## Configuration File
`rfm` will create a configuration file in the user's home directory containing connection parameters for all devices (selectable by `-device` options) the user has ever specified.
This means after connecting succesfully once to a new device this can always be reaccessed by just providing the chosen name to `-device` without the need to reenter `-domain`, `-port` and/or `-password`.

To create or update settings just specify `-device` and the parameters you want to set or update.

### Example
```
# Create a new configuration for "first_device". This will be saved in ~/rfm.toml
rfm ls -device first_device -domain some.domain -port 1234 0:/

# Create a new configuration for "second_device"
rfm ls -device second_device -domain some.other.domain -port 2345 0:/

# Update configuration of "first_device" after enabling a non-default password
rfm ls -device first_device -password my_insecure_password 0:/

# Use second_device
rfm ls -device second_device 0:/
```

## Feedback
Please provide any feedback either here in the Issues or send a pull request or go to [the Duet3D forum](https://forum.duet3d.com/topic/10880).
