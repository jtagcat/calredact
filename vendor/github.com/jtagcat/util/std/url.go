package std

import (
	"net/url"
	"path"
	"strings"
)

// k8s.io/helm/pkg/urlutil
// mod: supports // and /
// in case of multiple absolute paths, last is used
func URLJoin(baseURL string, paths ...string) (string, error) {
	// mod:
	// base is replaced by first with //
	newBase := -1
	for i, p := range paths {
		if strings.HasPrefix(p, "//") {
			newBase = i
		}
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	if newBase > -1 {
		old := u
		u, err = url.Parse(paths[newBase])
		if err != nil {
			return "", err
		}

		u.Scheme = old.Scheme
		if u.User == nil {
			u.User = old.User
		}

		paths = paths[newBase+1:]
	}

	// mod:
	// allow rooting to domain with /
	absPath := -1
	for i, p := range paths {
		if strings.HasPrefix(p, "/") {
			absPath = i
		}
	}

	// We want path instead of filepath because path always uses /.
	if absPath > -1 {
		u.Path = path.Join(paths[absPath:]...)
	} else {
		all := []string{u.Path}
		all = append(all, paths...)
		u.Path = path.Join(all...)
	}

	return u.String(), nil
}

func URLJoinNoErr(baseURL string, paths ...string) string {
	joined, _ := URLJoin(baseURL, paths...)

	return joined
}
