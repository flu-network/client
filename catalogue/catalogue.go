package catalogue

import (
	"path/filepath"
	"sync"
)

// Cat is a wrapper around the on-disk catalogue data for Flu clients. There should only be one
// cat per physical computer, and a single process accessing the cat files at any time. A cat
// consists of an index file (index.json) and several progress files (each named after the sha1
// hash of the files whose progress is being tracked).
// Most file-system access should be performed with a mutex lock. Public methods simply acquire
// the lock and then call private methods that assume the lock exists.
type Cat struct {
	DataDir   string
	indexFile *IndexFile
	lock      sync.Mutex
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
// data. Returns a descriptive error if unable to acquire a lock. This should be called before
// invoking any other methods on Cat. Note, this lock is different from a mutex. Our user-space
// lock indicates that the process is unique, not that this thread has exclusive access to some
// resource.
func (c *Cat) Init() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.indexFile = &IndexFile{}
	err := c.indexFile.Init(c.DataDir)
	if err != nil {
		return err
	}
	return nil
}

// ShareFile generates an IndexRecord for the given filepath (unless an identical file has
// already been shared) and refreshes the inderlying IndexFile. Sharing a file assumes that the
// file has been downloaded completely.
func (c *Cat) ShareFile(path string) (*IndexRecord, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	record, err := generateIndexRecordForFile(path)
	if err != nil {
		return nil, err
	}

	err = c.indexFile.AddIndexRecord(record)
	if err != nil {
		return nil, err
	}

	record.ProgressFile = NewProgressFile(record, c.DataDir)
	record.ProgressFile.save()

	return record, nil
}
