package std

func Deduplicate[T comparable](sl []T) (unique []T) {
	slMap := make(map[T]struct{})

	for _, comp := range sl {
		if _, ok := slMap[comp]; !ok {
			slMap[comp] = struct{}{}
			unique = append(unique, comp)
		}
	}

	return
}
