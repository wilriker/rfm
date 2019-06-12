package rfm

import "strings"

/// Excludes contains simple starts-with patterns for upload/download excludes
type Excludes struct {
	excls []string
}

func (e *Excludes) String() string {
	return strings.Join(e.excls, ",")
}

func (e *Excludes) Set(value string) error {
	e.excls = append(e.excls, value)
	return nil
}

// ForEach performs the given function on all entries
func (e *Excludes) ForEach(f func(string) string) {
	for i := 0; i < len(e.excls); i++ {
		e.excls[i] = f(e.excls[i])
	}
}

// Contains checks if the given path starts with any of the known excludes
func (e *Excludes) Contains(path string) bool {
	for _, excl := range e.excls {
		if strings.HasPrefix(path, excl) {
			return true
		}
	}
	return false
}
