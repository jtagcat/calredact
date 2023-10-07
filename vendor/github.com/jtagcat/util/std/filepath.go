package std

import "path/filepath"

// Either base is returned or base + elem, never outside base,
// except when using symlinks
func SafeJoin(base string, elem ...string) string {
	for _, e := range elem {
		e = filepath.Join("/", e)
		if e == "/" {
			continue
		}

		base = filepath.Join(base, e)
	}

	return base
}
