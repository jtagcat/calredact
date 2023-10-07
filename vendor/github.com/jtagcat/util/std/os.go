package std

import "os"

// os.ReadFile, but returns string
func ReadFile(name string) (string, error) {
	b, err := os.ReadFile(name)
	return string(b), err
}
