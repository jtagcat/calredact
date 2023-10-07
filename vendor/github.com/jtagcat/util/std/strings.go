package std

import (
	"strings"
)

// strings.Cut, but starting from last character, found is either empty or seperator
// Deprecated: use RevCutv2, when !found, rightOf is returned
func RevCutv1(s, sep string) (leftOf, rightOf string, found bool) {
	leftOf, rightOf, found = RevCutv2(s, sep)
	if found {
		return
	}

	return rightOf, "", false
}

func RevCutv2(s, sep string) (leftOf, rightOf string, found bool) {
	if i := strings.LastIndex(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return "", s, false
}

func CutFunc(s string, f func(rune) bool) (leftOf, rightOf string, found bool) {
	for i, r := range s {
		if f(r) {
			return s[:i], s[i+1:], true
		}
	}

	return s, "", false
}

func TrimLen(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}

	return string(r[:max])
}

// always returns [n]string; discards subelements after n
func StableSplitN(s, sep string, n int) []string {
	splitted := strings.SplitN(s, sep, n)
	if len(splitted) == n {
		return splitted
	}

	stable := make([]string, n)
	_ = copy(stable, splitted)

	return stable
}
