package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SemverParts represents the parts of a semantic version.
type SemverParts struct {
	Major int
	Minor int
	Patch int
	Extra string
}

// parseSemver parses a semantic version string and returns its parts.
func ParseSemver(semver string) (SemverParts, error) {
	// Regex to match semver (basic validation)
	pattern := `^v?(\d+)(\.\d+)?(\.\d+)?(-.+)?$`
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(semver)
	if matches == nil {
		return SemverParts{}, fmt.Errorf("invalid semver: %s", semver)
	}

	// Convert Major, Minor, and Patch to integers
	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return SemverParts{}, fmt.Errorf("invalid major version: %s", matches[1])
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		minor = 0
		// return SemverParts{}, fmt.Errorf("invalid minor version: %s", matches[2])
	}
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		patch = 0
		// return SemverParts{}, fmt.Errorf("invalid patch version: %s", matches[3])
	}

	parts := SemverParts{
		Major: major,
		Minor: minor,
		Patch: patch,
	}

	// Extra part (e.g., pre-release or build metadata)
	if len(matches) > 4 {
		parts.Extra = strings.TrimPrefix(matches[4], "-")
	}

	return parts, nil
}
