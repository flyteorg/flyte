package utils

func CopyMap(o map[string]string) (r map[string]string) {
	if o == nil {
		return nil
	}
	r = make(map[string]string, len(o))
	for k, v := range o {
		r[k] = v
	}
	return
}

func Contains(s []string, e string) bool {
	if s == nil {
		return false
	}

	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
