package utils

import "regexp"

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

func GetSanitizedPrometheusKey(key string) (string, error) {
	// Make a Regex to say we only want letters, numbers, : and _
	reg, err := regexp.Compile("[^a-zA-Z0-9_:]*")
	if err != nil {
		return "", err
	}
	return reg.ReplaceAllString(key, ""), nil
}
