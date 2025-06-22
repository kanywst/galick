// Package cli provides the command-line interface for galick.
package cli

import (
	"strings"

	"github.com/kanywst/galick/internal/constants"
)

// parsePushLabels parses a comma-separated list of key=value pairs into a map.
// e.g., "instance=ci-run,build_number=123" becomes {"instance": "ci-run", "build_number": "123"}
func parsePushLabels(labels string) map[string]string {
	result := make(map[string]string)
	if labels == "" {
		return result
	}

	pairs := strings.Split(labels, ",")
	for _, pair := range pairs {
		// Try splitting by "=" first
		kv := strings.SplitN(pair, "=", constants.DefaultSplitParts)
		if len(kv) != constants.DefaultSplitParts {
			// If no "=" found, try ":" as separator
			kv = strings.SplitN(pair, ":", constants.DefaultSplitParts)
		}

		if len(kv) == constants.DefaultSplitParts {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			if key != "" {
				result[key] = value
			}
		}
	}

	return result
}
