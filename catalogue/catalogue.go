package catalogue

import (
	"path/filepath"
)

// Cat is a wrapper around the on-disk catalogue data for Flu clients. There should only be one
// cat per physical computer, and a single process accessing the cat files at any time. A cat
// consists of an index file (index.json) and several progress files (each named after the sha1
// hash of the files whose progress is being tracked).
type Cat struct {
	DataDir   string
	indexFile *IndexFile
}

// NewCat returns a Cat struct, initialized to the given data directory
func NewCat(dir string) (*Cat, error) {
	cleanPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	return &Cat{
		DataDir:   cleanPath,
		indexFile: nil,
	}, nil
}

// Init initializes or attempts to acquire an exclusive user-space lock on the on-disk catalogue
// data. Returns a descriptive error if unable to acquire a lock.
func (c *Cat) Init() error {
	c.indexFile = &IndexFile{}
	err := c.indexFile.Init(c.DataDir)
	if err != nil {
		return err
	}

	return nil
}
