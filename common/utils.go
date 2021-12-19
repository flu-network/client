package common

import (
	"bufio"
	"crypto/sha1"
	"os"
)

// HashFile returns a 20-byte slice representing the sha1 hash of the file's contents or an error.
// TODO: come up with some sort of async/progress-bar method because at 500mbps of disk throughput
// this could take 10 seconds for a 5GB file, and 1.5 minutes for a 50GB file.
func HashFile(path string) (*Sha1Hash, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	s, h := bufio.NewScanner(f), sha1.New()
	for s.Scan() {
		h.Write(s.Bytes())
	}

	return (&Sha1Hash{}).FromSlice(h.Sum(nil)), nil
}
