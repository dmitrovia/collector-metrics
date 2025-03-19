// Package random validate functions
// working with random validation.
package validate

import (
	"fmt"
	"net/http"
	"regexp"
)

// IsMatchesTemplate - checks
// for regular expression matches.
func IsMatchesTemplate(
	addr string,
	pattern string,
) (bool, error) {
	res, err := matchString(pattern, addr)
	if err != nil {
		return false, err
	}

	return res, err
}

func matchString(pattern string, s string) (bool, error) {
	re, err := regexp.Compile(pattern)
	if err == nil {
		return re.MatchString(s), nil
	}

	return false, fmt.Errorf("MatchString: %w", err)
}

// IsMethodPost - checks that
// the method meets the post requirements.
func IsMethodPost(method string) bool {
	return method == http.MethodPost
}
