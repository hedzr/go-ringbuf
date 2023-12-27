//go:build verbose
// +build verbose

package state

import "log/slog"

const VerboseEnabled = true

func Verbose(msg string, args ...any) {
	slog.Info(msg, args...)
}
