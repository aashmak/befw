package server

func contentEncodingContains(a []string, x string) bool {
	for _, s := range a {
		if s == x {
			return true
		}
	}

	return false
}
