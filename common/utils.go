package common

import (
	"crypto/sha1"
	"fmt"
	"io"
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
	defer f.Close()

	hash := sha1.New()
	hashBuffer := make([]byte, 4096) // A reasonably-recent mac's block size

	for {
		bytesRead, err := f.Read(hashBuffer)
		if bytesRead > 0 {
			wrote, err := hash.Write(hashBuffer[:bytesRead])
			if err != nil {
				return nil, err
			}
			if wrote != bytesRead {
				return nil, fmt.Errorf("internal error writing to hash")
			}
		}

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		// else do nothing
	}

	return (&Sha1Hash{}).FromSlice(hash.Sum(nil)), nil
}

func failHard(err error) {
	if err != nil {
		panic(err)
	}
}
