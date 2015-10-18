package util

import "os"

// EnsureFileRemoved checks the given path and tries to Close and Remove it if
// it exists.
func EnsureFileRemoved(name string) error {
	if f, err := os.Open(name); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if err = f.Close(); err != nil {
			return err
		}
		if err = os.Remove(name); err != nil {
			return err
		}
	}

	return nil
}
