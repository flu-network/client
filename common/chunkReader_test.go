package common

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

const sectionSize = int64(1 << 22) // 4mb in bytes

func TestChunkReaderHashMatching(t *testing.T) {
	type testcase struct {
		fileSize int
		offset   int64
	}

	testcases := []testcase{
		{
			fileSize: 0,
			offset:   0,
		},
		{
			fileSize: 3,
			offset:   0,
		},
		{
			fileSize: 1600, // arbitrary big number < sectionSize
			offset:   0,
		},
		// TODO: add coverage for offset > 0 cases
	}

	for _, tc := range testcases {
		t.Run("Hashes match", func(t *testing.T) {
			testFilePath := fmt.Sprintf("/tmp/flu_abc_%d.txt", tc.fileSize)
			genStableRandomishData(tc.fileSize, testFilePath)

			hash, err := HashFile(testFilePath)
			failHard(err)

			fd, err := os.Open(testFilePath)
			failHard(err)
			chunkReader := NewChunkReader(io.NewSectionReader(fd, 0, sectionSize))

			if hash.Data != chunkReader.Hash.Data {
				t.Fatalf(
					"Expected hashes original: %v and reader: %v to match",
					hash.String(),
					chunkReader.Hash.String(),
				)
			}
		})
	}

}

// genStableRandomishData writes sizeInBytes characters from the set [a-z] to a file at filePath. If
// filePath already exists it will overwrite the file. The characters are written in order so that
// calls to the function with the same parameters will produce the same result
func genStableRandomishData(sizeInBytes int, filePath string) {
	chars := "abcdefghijklmnopqrstuvwxyz"
	data := make([]byte, sizeInBytes)
	for i := 0; i < sizeInBytes; i++ {
		data[i] = chars[i%26]
	}
	dest := filepath.Clean(filePath)
	failHard(os.WriteFile(dest, data, 0777))
}
