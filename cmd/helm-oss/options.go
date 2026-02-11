package main

import (
	"time"
)

// options represents global command options (global flags).
type options struct {
	timeout time.Duration
	verbose bool
}

// newDefaultOptions returns default options.
func newDefaultOptions() *options {
	return &options{
		timeout: 5 * time.Minute,
		verbose: false,
	}
}
