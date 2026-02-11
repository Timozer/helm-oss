package helmutil

import (
	"io"

	"helm.sh/helm/v3/pkg/provenance"
)

// Digest hashes a reader and returns a SHA256 digest.
func Digest(in io.Reader) (string, error) {
	return provenance.Digest(in)
}

// DigestFile calculates a SHA256 hash for a given file.
func DigestFile(filename string) (string, error) {
	return provenance.DigestFile(filename)
}
