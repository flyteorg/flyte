package utils

import (
	"net/url"
	"regexp"
)

// A utility function to be used in tests. It parses urlString as url.URL or panics if it's invalid.
func MustParseURL(urlString string) url.URL {
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}

	return *u
}

// A utility function to be used in tests. It returns the address of the passed value.
func RefInt(val int) *int {
	return &val
}

// A utility function to be used in tests. It compiles regexpString as regexp.Regexp or panics if it's invalid.
func MustCompileRegexp(regexpString string) regexp.Regexp {
	r, err := regexp.Compile(regexpString)
	if err != nil {
		panic(err)
	}

	return *r
}
