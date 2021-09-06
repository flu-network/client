package catalogue

import (
	"fmt"
	"os"
	"path/filepath"
)

const indexFileName = "index.json"

// Cat is a wrapper around the on-disk catalogue data for Flu clients. There should only be one
// cat per physical computer, and a single process accessing the cat files at any time. A cat
// consists of an index file (index.json) and several progress files (each named after the sha1
// hash of the files whose progress is being tracked).
type Cat struct {
	DataDir   string
	IndexFile *IndexFile
}

// NewCat returns a Cat struct, initialized to the given data directory
func NewCat(dir string) (*Cat, error) {
	cleanPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	return &Cat{
		DataDir:   cleanPath,
		IndexFile: nil,
	}, nil
}

// Init initializes or attempts to acquire an exclusive user-space lock on the on-disk catalogue
// data. Returns a descriptive error if unable to acquire a lock.
func (c *Cat) Init() error {
	// ensure directory exists
	if err := os.MkdirAll(c.DataDir, os.ModePerm); err != nil {
		return err
	}

	// check the container is accessible and is a directory
	fileInfo, err := os.Stat(c.DataDir)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("DataDir '%s' is not a directory", c.DataDir)
	}

	// ensure the file exists
	// TODO: Complete
	indexFilePath := filepath.Join(c.DataDir, indexFileName)
	if _, err := os.Stat(indexFilePath); os.IsNotExist(err) {
		// create it if it doesn't exist
		// indexFile := IndexFile{
		// 	pid:         os.Getpid(),
		// 	lastTouched: time.Now().Unix(),
		// 	index:       map[[20]byte]IndexRecord{},
		// }
	}

	// ensure exclusivity
	// check the PIDs match, failing which, check lastTouched was more than 30 seconds ago
	return nil
}
