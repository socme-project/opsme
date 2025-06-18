package opsme

import (
	"errors"
	"os"
	"strings"
)

// GetKeyFromFile reads a key from a specified file path.
// It returns the key as a byte slice and an error if any issues occur.
func GetKeyFromFile(path string) (key []byte, err error) {
	path = strings.TrimSpace(path)
	if path == "" {
		err = errors.New("path cannot be empty")
		return
	}

	key, err = os.ReadFile(path)
	if err != nil {
		return
	}
	if len(key) == 0 {
		err = errors.New("key file is empty")
		return
	}
	return
}

// WithKnownHostsPath sets the path for the known_hosts file.
func (op *Operator) WithKnownHostsPath(path string) *Operator {
	op.KnownHostsPath = path
	return op
}
