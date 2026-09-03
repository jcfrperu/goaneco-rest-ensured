package rest

import (
	"net/url"
	"strings"
)

// GetPath extracts only the path component from a URI string, stripping query parameters.
// Relative paths are normalized to start with "/".
// Mirrors Java's PathSupport.getPath behavior.
func GetPath(uri string) string {
	if isFullyQualifiedURI(uri) {
		parsed, err := url.Parse(uri)
		if err != nil || parsed.Path == "" {
			return "/"
		}
		return parsed.Path
	}
	// Relative path: strip everything from the first "?" onward.
	path := uri
	if idx := strings.Index(uri, "?"); idx != -1 {
		path = uri[:idx]
	}
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

// isFullyQualifiedURI returns true when the string begins with a valid URI scheme
// (e.g. "http://", "https://"). A "://" that appears only inside a query-parameter
// value is not treated as a scheme because the scheme characters before it include
// "/" or "?" which are not valid scheme characters.
func isFullyQualifiedURI(uri string) bool {
	idx := strings.Index(uri, "://")
	if idx <= 0 {
		return false
	}
	scheme := uri[:idx]
	for i, c := range scheme {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
			// always valid
		case c >= '0' && c <= '9', c == '+', c == '-', c == '.':
			if i == 0 {
				return false // first char must be a letter
			}
		default:
			return false
		}
	}
	return true
}
