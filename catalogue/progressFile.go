package catalogue

import (
	"os"
	"path/filepath"

	"github.com/flu-network/client/common/bitset"
)

// progressFile is an in-memory representation of a file on disk containing a bitset, which shows
// which 'chunks' of a file has been downloaded. If the entire file has been downloaded, all bits
// in the set are 'on'. Serialization and deserialization methods assume the caller has already
// obtained a mutex, to avoid writing while reading
type progressFile struct {
	progress bitset.Bitset
	filePath string
}

// Full returns true if all items between 0:size are set to true, and false if not.
func (p *progressFile) Full() bool {
	return p.progress.Full()
}

// Size returns the number of elements in the bitset, both set and unset. Size is always
// non-negative
func (p *progressFile) Size() int {
	return p.progress.Size()
}

// Count returns the number of elements in the bitset that are set (true)
func (p *progressFile) Count() int {
	return p.progress.Count()
}

// Overlap takes a sorted  list of ranges ([]uint16 of even length) where each consecutive pair of
// numbers represents an inclusive range [start, end] of bits. The return value is a sorted list of
// non-overallping ranges that are set to 'on' within the underlying bitset
func (p *progressFile) Overlap(ranges []uint16) []uint16 {
	return p.progress.Overlap(ranges)
}

// Ranges returns a sorted, non-overallping list of ranges of the underlying bitset that are set to
// true.
func (p *progressFile) Ranges() []uint16 {
	return p.progress.Ranges()
}

// newProgressFile returns a new progressFile for the given IndexRecord, assuming the IndexRecord
// is intact and preset in full.
func newProgressFile(record *indexRecord, dataDir string) *progressFile {
	sizeInBytes := int(record.SizeInBytes)
	chunkCount := sizeInBytes / record.ChunkSize
	if sizeInBytes%record.ChunkSize != 0 {
		chunkCount++
	}
	set := *bitset.NewBitset(int(chunkCount))
	return &progressFile{
		progress: set,
		filePath: filepath.Join(dataDir, record.Sha1Hash.String()),
	}
}

func (p *progressFile) save() error {
	data := p.progress.Serialize()
	progressFilePath := filepath.Join(p.filePath)
	err := os.WriteFile(progressFilePath, data, 0664)
	if err != nil {
		return err
	}

	return nil
}

func (p *progressFile) delete() error {
	return os.Remove(filepath.Join(p.filePath))
}

// deserializeProgressFile reads bytes on disk into an in-memory progressFile.
func deserializeProgressFile(record *indexRecord, dataDir string) (*progressFile, error) {
	progressFilePath := filepath.Join(dataDir, record.Sha1Hash.String())

	data, err := os.ReadFile(progressFilePath)
	if err != nil {
		return nil, err
	}

	set, err := bitset.Deserialize(data)
	if err != nil {
		return nil, err
	}

	return &progressFile{
		progress: *set,
		filePath: progressFilePath,
	}, nil
}
