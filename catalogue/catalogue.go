package catalogue

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/flu-network/client/common"
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
	record.ProgressFile.Progress.Fill()
	err = record.ProgressFile.save()
	if err != nil {
		return nil, err
	}

	return record, nil
}

// UnshareFile immediately deletes all references to it from flu's index. Any transfers in progress
// will throw errors and stop. The actual file is not affected in any way.
func (c *Cat) UnshareFile(ir *IndexRecord) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	err := ir.ProgressFile.delete()
	if err != nil {
		return err
	}

	err = c.indexFile.RemoveIndexRecord(ir)
	if err != nil {
		return err
	}

	return nil
}

// RegisterDownload creates a record of the download in flu's index. This is identical to
// c.ShareFile except that the progress file will register an empty bitset.
func (c *Cat) RegisterDownload(
	sizeInBytes uint64,
	chunkCount uint32,
	chunkSizeInBytes uint32,
	sha1Hash *common.Sha1Hash,
	filename string,
) error {
	indexRecord := IndexRecord{
		FilePath:     fmt.Sprintf("~/Downloads/%s", filename),
		SizeInBytes:  int64(sizeInBytes),
		Sha1Hash:     *sha1Hash,
		ProgressFile: nil,
		ChunkSize:    int(chunkSizeInBytes),
	}

	err := c.indexFile.AddIndexRecord(&indexRecord)
	if err != nil {
		return err
	}

	indexRecord.ProgressFile = NewProgressFile(&indexRecord, c.DataDir)
	indexRecord.ProgressFile.save()
	if err != nil {
		return err
	}

	return nil
}

// ListFiles lists the files that exist in the catalogue. Not all indexed files have been downloaded
// in their entirety. The result is a deep copy of the underlying catalogue data, so mutating it is
// okay.
func (c *Cat) ListFiles() ([]IndexRecord, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	result := make([]IndexRecord, 0, len(c.indexFile.index))

	for _, rec := range c.indexFile.index {
		fmt.Println(rec.FilePath)
		if rec.ProgressFile == nil {
			p, err := DeserializeProgressFile(&rec, c.DataDir)
			if err != nil {
				return nil, err
			}
			rec.ProgressFile = p
		}
		result = append(result, rec)
	}

	return result, nil
}

// Rehash attempts to recalculate the hash for a given indexRecord. If it fails, a blank hash and an
// error are returned.
func (c *Cat) Rehash(ir *IndexRecord) (*common.Sha1Hash, error) {
	currentHash, err := hashFile(ir.FilePath)
	if err != nil {
		return (&common.Sha1Hash{}).Blank(), err
	}
	return currentHash, nil
}

// Contains returns the indexRecord of the file specified by the hash, or an error if the file
// cannot be accessed for any reason.
func (c *Cat) Contains(hash *common.Sha1Hash) (*IndexRecord, error) {
	if record, found := c.indexFile.index[*hash]; found {
		result, err := c.fill(&record)
		return result, err
	}
	return nil, fmt.Errorf("file not found")
}

// Get is the same as contains, except that if the record does not exist it will panic.
func (c *Cat) Get(hash *common.Sha1Hash) *IndexRecord {
	result, err := c.Contains(hash)
	if err != nil {
		panic(err)
	}
	return result
}

func (c *Cat) GetChunkReader(ir *IndexRecord, chunk int64) (*common.ChunkReader, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return ir.getChunkReader(chunk)
}

func (c *Cat) fill(rec *IndexRecord) (*IndexRecord, error) {
	if rec.ProgressFile == nil {
		p, err := DeserializeProgressFile(rec, c.DataDir)
		if err != nil {
			return rec, err
		}
		rec.ProgressFile = p
	}
	return rec, nil
}
