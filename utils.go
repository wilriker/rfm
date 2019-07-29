package rfm

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	Byte = 1 << (10 * iota)
	Kilobyte
	Megabyte
	Gigabyte
	Terabyte
	paddedWidth  = 5
	defaultMount = "0:/"
)

var multiSlashRegex = regexp.MustCompile(`/{2,}`)
var absRemotePath = regexp.MustCompile(`^[0-9]:/`)

// CleanRemotePath will reduce multiple consecutive slashes into one and
// then remove a trailing slash if any.
func CleanRemotePath(path string) string {
	cleanedPath := strings.TrimSpace(path)
	cleanedPath = strings.TrimSuffix(cleanedPath, "/")
	if !absRemotePath.MatchString(cleanedPath) {
		cleanedPath = defaultMount + cleanedPath
	}
	cleanedPath = multiSlashRegex.ReplaceAllString(cleanedPath, "/")
	return cleanedPath
}

// GetAbsPath tries to make an absolute path from the given value
// in case of an error it returns the original value unchanged.
func GetAbsPath(path string) string {

	// Get absolute path from user's input
	absPath, err := filepath.Abs(path)
	if err == nil {
		return absPath
	}
	return path
}

// HumanReadableSize returns a human-readable byte string of the form 10M, 12.5K, and so forth.  The following units are available:
//	T: Terabyte
//	G: Gigabyte
//	M: Megabyte
//	K: Kilobyte
//	B: Byte
// The unit that results in the smallest number greater than or equal to 1 is always chosen.
func HumanReadableSize(bytes uint64) string {
	unit := ""
	value := float64(bytes)

	switch {
	case bytes >= Terabyte:
		unit = "T"
		value = value / Terabyte
	case bytes >= Gigabyte:
		unit = "G"
		value = value / Gigabyte
	case bytes >= Megabyte:
		unit = "M"
		value = value / Megabyte
	case bytes >= Kilobyte:
		unit = "K"
		value = value / Kilobyte
	case bytes >= Byte:
		unit = "B"
	case bytes == 0:
		return fmt.Sprintf("%*d", paddedWidth+1, 0)
	}

	result := fmt.Sprintf("%*.1f", paddedWidth, value)
	if strings.HasSuffix(result, ".0") {
		result = "  " + strings.TrimSuffix(result, ".0")
	}
	return result + unit
}
