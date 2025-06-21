// Package constants provides common constants used throughout the galick application.
package constants

// File permission constants
const (
	// DirPermissionDefault is the default permission for directories (0755)
	DirPermissionDefault = 0o755
	// FilePermissionDefault is the default permission for files (0644)
	FilePermissionDefault = 0o644
	// FilePermissionPrivate is the permission for private files (0600)
	FilePermissionPrivate = 0o600
)

// Time unit conversion constants
const (
	// NanoToMillisecond is the conversion factor from nanoseconds to milliseconds (10^6)
	NanoToMillisecond = 1000000
	// NanoToSecond is the conversion factor from nanoseconds to seconds (10^9)
	NanoToSecond = 1000000000
)

// Common numeric constants
const (
	// Percentage100 represents 100% for percentage calculations
	Percentage100 = 100
	// MilliToSec represents the conversion from milliseconds to seconds (1000)
	MilliToSec = 1000
	// DefaultSplitParts is the default number of parts for string splitting
	DefaultSplitParts = 2
)
