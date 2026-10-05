package auth

import (
	"strings"
)

// MatchesIdentity reports whether the caller's subject, app id, or email is in want.
func MatchesIdentity(identity IdentityContext, want []string) bool {
	for _, candidate := range want {
		if candidate == "" {
			continue
		}
		if candidate == identity.UserID() || candidate == identity.AppID() ||
			candidate == identity.UserInfo().GetEmail() {
			return true
		}
	}
	return false
}

// HasGroup reports whether the named claim holds one of want, while ignoring a leading "/".
func HasGroup(identity IdentityContext, claim string, want []string) bool {
	if claim == "" {
		return false
	}
	for _, group := range ClaimStrings(identity.Claims()[claim]) {
		for _, candidate := range want {
			if candidate != "" && strings.TrimPrefix(group, "/") == strings.TrimPrefix(candidate, "/") {
				return true
			}
		}
	}
	return false
}

// ClaimStrings reads a claim that may arrive as a list or as a single string.
func ClaimStrings(claim interface{}) []string {
	switch value := claim.(type) {
	case []string:
		return value
	case string:
		return []string{value}
	case []interface{}:
		groups := make([]string, 0, len(value))
		for _, item := range value {
			if group, ok := item.(string); ok {
				groups = append(groups, group)
			}
		}
		return groups
	}
	return nil
}
