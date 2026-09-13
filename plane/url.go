package plane

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const apiPath = "/api/v1"

func normalizeBaseURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = DefaultBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || strings.Contains(raw, "#") || u.Scheme == "" || u.Host == "" || u.Opaque != "" || u.User != nil ||
		u.Fragment != "" || u.RawQuery != "" || u.ForceQuery || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("invalid Plane base URL")
	}

	basePath := strings.TrimRight(u.Path, "/")
	if !strings.HasSuffix(basePath, apiPath) {
		basePath += apiPath
	}
	u.Path = basePath + "/"
	u.RawPath = ""
	u.RawQuery = ""
	u.ForceQuery = false
	return u, nil
}

// NormalizeBaseURL returns a canonical Plane API base with exactly one
// trailing /api/v1/ path segment.
func NormalizeBaseURL(raw string) (string, error) {
	u, err := normalizeBaseURL(raw)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// EscapePathSegment escapes one dynamic route segment without allowing its
// value to introduce a slash or query delimiter.
func EscapePathSegment(value string) string {
	return url.PathEscape(value)
}

func copyValues(values url.Values) url.Values {
	if values == nil {
		return make(url.Values)
	}
	clone := make(url.Values, len(values))
	for key, entries := range values {
		clone[key] = append([]string(nil), entries...)
	}
	return clone
}

func validatePerPageValue(value string) error {
	perPage, err := strconv.Atoi(value)
	if err != nil || perPage < 1 || perPage > 100 {
		return fmt.Errorf("per_page must be between 1 and 100")
	}
	return nil
}

func validatePerPageQuery(values url.Values) error {
	for _, value := range values["per_page"] {
		if err := validatePerPageValue(value); err != nil {
			return err
		}
	}
	return nil
}

// WithPagination returns a copied query containing the shared pagination
// controls. Existing resource-specific parameters are retained.
func WithPagination(values url.Values, options PaginationOptions) (url.Values, error) {
	query := copyValues(values)
	if options.PerPage != 0 {
		if options.PerPage < 1 || options.PerPage > 100 {
			return nil, fmt.Errorf("per_page must be between 1 and 100")
		}
		query.Set("per_page", strconvItoa(options.PerPage))
	}
	if options.Cursor != "" {
		query.Set("cursor", options.Cursor)
	}
	if options.Fields != "" {
		query.Set("fields", options.Fields)
	}
	if options.Expand != "" {
		query.Set("expand", options.Expand)
	}
	if err := validatePerPageQuery(query); err != nil {
		return nil, err
	}
	return query, nil
}

// PaginationQuery is a descriptive alias for WithPagination.
func PaginationQuery(values url.Values, options PaginationOptions) (url.Values, error) {
	return WithPagination(values, options)
}

func routePath(route string) (string, string, error) {
	if strings.Contains(route, "#") {
		return "", "", fmt.Errorf("invalid Plane route")
	}
	u, err := url.Parse(route)
	if err != nil || u.IsAbs() || u.Host != "" || u.User != nil || u.Opaque != "" ||
		u.Fragment != "" || u.RawQuery != "" || u.ForceQuery {
		return "", "", fmt.Errorf("invalid Plane route")
	}
	decodedPath := u.Path
	for _, segment := range strings.Split(decodedPath, "/") {
		if segment == "." || segment == ".." {
			return "", "", fmt.Errorf("invalid Plane route")
		}
	}
	return decodedPath, u.EscapedPath(), nil
}

func joinRoute(base *url.URL, route string) (*url.URL, error) {
	decodedRoute, escapedRoute, err := routePath(route)
	if err != nil {
		return nil, err
	}
	if decodedRoute == apiPath || strings.HasPrefix(decodedRoute, apiPath+"/") {
		decodedRoute = strings.TrimPrefix(decodedRoute, apiPath)
		escapedRoute = strings.TrimPrefix(escapedRoute, apiPath)
	}

	joined := *base
	basePath := strings.TrimSuffix(base.Path, "/")
	baseEscapedPath := strings.TrimSuffix(base.EscapedPath(), "/")
	routePath := strings.TrimPrefix(decodedRoute, "/")
	routeEscapedPath := strings.TrimPrefix(escapedRoute, "/")
	if routePath == "" {
		joined.Path = basePath + "/"
		joined.RawPath = baseEscapedPath + "/"
	} else {
		joined.Path = basePath + "/" + routePath
		joined.RawPath = baseEscapedPath + "/" + routeEscapedPath
	}
	return &joined, nil
}

// BuildURL normalizes the base, safely joins an explicit route path, and
// encodes query values using url.Values. Routes may include an optional
// /api/v1 prefix for convenience, but no route is inferred from a resource
// name or documentation slug.
func BuildURL(baseURL, route string, values url.Values) (string, error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return "", err
	}
	if err := validatePerPageQuery(values); err != nil {
		return "", err
	}
	joined, err := joinRoute(base, route)
	if err != nil {
		return "", err
	}
	joined.RawQuery = copyValues(values).Encode()
	return joined.String(), nil
}

func strconvItoa(value int) string {
	return strconv.Itoa(value)
}
