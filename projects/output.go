package projects

import "planeshift/objects"

// RawOutput adapts a typed Project response or page to the repository's
// centralized output dispatcher without importing config or implementing a
// format switch in the resource package.
func RawOutput(value any) (objects.RawJSON, error) {
	return objects.NewRawJSONFromValue(value)
}
