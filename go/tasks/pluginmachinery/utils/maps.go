package utils

// This function unions a list of maps (each can be nil or populated) by allocating a new map.
// Conflicting keys will always defer to the later input map's corresponding value.
func UnionMaps(maps ...map[string]string) map[string]string {
	size := 0
	for _, m := range maps {
		size += len(m)
	}

	composite := make(map[string]string, size)
	for _, m := range maps {
		for k, v := range m {
			composite[k] = v
		}
	}

	return composite
}
