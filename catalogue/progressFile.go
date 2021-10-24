package catalogue

import (
	"os"
	"path/filepath"

	"github.com/flu-network/client/common/bitset"
)

// ProgressFile is an in-memory representation of a file on disk containing a bitset, which shows
// which 'chunks' of a file has been downloaded. If the entire file has been downloaded, all bits
// in the set are 'on'. Serialization and deserialization methods assume the caller has already
// obtained a mutex, to avoid writing while reading
type ProgressFile struct {
	Progress bitset.Bitset
	FilePath string
}

// NewProgressFile returns a new progressFile for the given IndexRecord, assuming the IndexRecord
// is intact and preset in full.
func NewProgressFile(record *IndexRecord, dataDir string) *ProgressFile {
	sizeInBytes := int(record.SizeInBytes)
	chunkCount := sizeInBytes / record.ChunkSize
	if sizeInBytes%record.ChunkSize != 0 {
		chunkCount++
	}
	set := *bitset.NewBitset(int(chunkCount))
	set.Fill()
	return &ProgressFile{
		Progress: set,
		FilePath: filepath.Join(dataDir, record.Sha1Hash.String()),
	}
}

func (p *ProgressFile) save() error {
	data := p.Progress.Serialize()
	progressFilePath := filepath.Join(p.FilePath)
	err := os.WriteFile(progressFilePath, data, 0664)
	if err != nil {
		return err
	}

	return nil
}

func (p *ProgressFile) delete() error {
	return os.Remove(filepath.Join(p.FilePath))
}

// DeserializeProgressFile reads bytes on disk into an in-memory progressFile.
func DeserializeProgressFile(record *IndexRecord, dataDir string) (*ProgressFile, error) {
	progressFilePath := filepath.Join(dataDir, record.Sha1Hash.String())

	data, err := os.ReadFile(progressFilePath)
	if err != nil {
		return nil, err
	}

	set, err := bitset.Deserialize(data)
	if err != nil {
		return nil, err
	}

	return &ProgressFile{
		Progress: *set,
		FilePath: progressFilePath,
	}, nil
}
