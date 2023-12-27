//go:build !verbose
// +build !verbose

package state

const VerboseEnabled = false

func Verbose(_ string, _ ...any) {}
