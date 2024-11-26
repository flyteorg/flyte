package filters

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var (
	InReg         = regexp.MustCompile(` in `)
	ContainsReg   = regexp.MustCompile(` contains `)
	InRegValue    = regexp.MustCompile(`(?s)\((.*)\)`)
	termOperators = []string{NotEquals, Equals, GreaterThanEquals, GreaterThan, LessThanEquals, LessThan, Contains, In}
)

// SplitTerms split the filter string and returns the map of strings
func SplitTerms(filter string) []string {
	if filter != "" {
		return strings.Split(filter, ",")
	}
	return []string{}
}

// Transform transform the field selector term from string to flyteadmin field selector syntax
func Transform(filters []string) (string, error) {
	adminFilter := ""
	for _, f := range filters {
		if lhs, op, rhs, ok := parse(f); ok {
			unescapedRHS, err := UnescapeValue(rhs)
			if err != nil {
				return "", err
			}
			if ok := validate(lhs, rhs); ok {
				transformFilter := transform(lhs, op, unescapedRHS)
				if len(adminFilter) > 0 {
					adminFilter = fmt.Sprintf("%v+%v", adminFilter, transformFilter)
				} else {
					adminFilter = fmt.Sprintf("%v", transformFilter)
				}
			} else {
				// TODO(Yuvraj): Add filters docs in error
				return "", fmt.Errorf("Please add a valid field selector")
			}
		}
	}
	return adminFilter, nil
}

// validate validate the field selector operation
func validate(lhs, rhs string) bool {
	// TODO Add Validation check with regular expression
	if len(lhs) > 0 && len(rhs) > 0 {
		return true
	}
	return false
}

// InvalidEscapeSequence indicates an error occurred unescaping a field selector
type InvalidEscapeSequence struct {
	sequence string
}

func (i InvalidEscapeSequence) Error() string {
	return fmt.Sprintf("invalid field selector: invalid escape sequence: %s", i.sequence)
}

// EscapeValue escapes strings to be used as values in filter queries.
func EscapeValue(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`,`, `\,`,
		`=`, `\=`,
	)
	return replacer.Replace(s)
}

// UnescapeValue unescapes a fieldSelector value and returns the original literal value.
// May return the original string if it contains no escaped or special characters.
func UnescapeValue(s string) (string, error) {
	// if there's no escaping or special characters, just return to avoid allocation
	if !strings.ContainsAny(s, `\,=`) {
		return s, nil
	}

	v := bytes.NewBuffer(make([]byte, 0, len(s)))
	inSlash := false
	for _, c := range s {
		if inSlash {
			switch c {
			case '\\', ',', '=':
				// omit the \ for recognized escape sequences
				v.WriteRune(c)
			default:
				// error on unrecognized escape sequences
				return "", InvalidEscapeSequence{sequence: string([]rune{'\\', c})}
			}
			inSlash = false
			continue
		}

		switch c {
		case '\\':
			inSlash = true
		case ',', '=':
			// unescaped , and = characters are not allowed in field selector values
			return "", UnescapedRune{r: c}
		default:
			v.WriteRune(c)
		}
	}

	// Ending with a single backslash is an invalid sequence
	if inSlash {
		return "", InvalidEscapeSequence{sequence: "\\"}
	}

	return v.String(), nil
}

// UnescapedRune indicates an error occurred unescaping a field selector
type UnescapedRune struct {
	r rune
}

func (i UnescapedRune) Error() string {
	return fmt.Sprintf("invalid field selector: unescaped character in value: %v", i.r)
}

// parse parse the filter string into an operation string and return the lhs,rhs value and operation type
func parse(filter string) (lhs, op, rhs string, ok bool) {
	for i := range filter {
		remaining := filter[i:]
		var results []string
		for _, op := range termOperators {
			if op == Contains {
				if ContainsReg.MatchString(filter) {
					results = ContainsReg.Split(filter, 2)
					return results[0], op, results[1], true
				}
			} else if op == In {
				if InReg.MatchString(filter) {
					results = InReg.Split(filter, 2)
					values := InRegValue.FindAllStringSubmatch(strings.TrimSpace(results[1]), -1)
					return results[0], op, values[0][1], true
				}
			} else {
				if strings.HasPrefix(remaining, op) {
					return filter[0:i], op, filter[i+len(op):], true
				}
			}
		}
	}
	return "", "", "", false
}

// transform it transform  the field selector operation and return flyteadmin filter syntax
func transform(lhs, op, rhs string) string {
	switch op {
	case GreaterThanEquals:
		return fmt.Sprintf("gte(%v,%v)", lhs, rhs)
	case LessThanEquals:
		return fmt.Sprintf("lte(%v,%v)", lhs, rhs)
	case GreaterThan:
		return fmt.Sprintf("gt(%v,%v)", lhs, rhs)
	case LessThan:
		return fmt.Sprintf("lt(%v,%v)", lhs, rhs)
	case Contains:
		return fmt.Sprintf("contains(%v,%v)", lhs, rhs)
	case NotEquals:
		return fmt.Sprintf("ne(%v,%v)", lhs, rhs)
	case Equals:
		return fmt.Sprintf("eq(%v,%v)", lhs, rhs)
	case In:
		return fmt.Sprintf("value_in(%v,%v)", lhs, rhs)
	}
	return ""
}
