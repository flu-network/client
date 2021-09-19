package catalogue

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
)

const defaultChunkSize = int(1 << 22) // 4mb in bytes

// IndexRecord describes a file that is 'known' by the flu client. The existence of an IndexRecord
// does not imply that the file exists locally. To find out which chunks of the file are
// downloaded, consult the progressFile. By convention, the progressFile is always named after
// the sha1 of the completely-downloaded file.
// All methods assume the caller has acquired a mutex granting exclusive access.
type IndexRecord struct {
	FilePath     string
	SizeInBytes  int64
	Sha1Hash     [20]byte
	ProgressFile *ProgressFile
	ChunkSize    int
}

func generateIndexRecordForFile(path string) (*IndexRecord, error) {
	cleanPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	fileStats, err := os.Stat(cleanPath)
	if err != nil {
		return nil, err
	}

	hash, err := hashFile(cleanPath)
	if err != nil {
		return nil, err
	}

	return &IndexRecord{
		FilePath:     cleanPath,
		SizeInBytes:  fileStats.Size(),
		Sha1Hash:     *hash,
		ProgressFile: nil,
		ChunkSize:    defaultChunkSize,
	}, nil

}

// hashFile returns a 20-array slice representing the sha1 hash of the file's contents or an error.
// TODO: come up with some sort of async/progress-bar method because at 500mbps of disk throughput
// this could take 10 seconds for a 5GB file, and 1.5 minutes for a 50GB file.
func hashFile(path string) (*[20]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	s, h := bufio.NewScanner(f), sha1.New()
	for s.Scan() {
		h.Write(s.Bytes())
	}

	hashArray := new([20]byte)
	copy(hashArray[:], h.Sum(nil))
	return hashArray, nil
}

// toJSON returns an indexRecordJSON, which can natively be marshalled into JSON
func (ir *IndexRecord) toJSON() *indexRecordJSON {
	return &indexRecordJSON{
		FilePath:    ir.FilePath,
		SizeInBytes: ir.SizeInBytes,
		Sha1Hash:    hex.EncodeToString(ir.Sha1Hash[:]),
		ChunkSize:   ir.ChunkSize,
	}
}

// UnmarshalJSON conforms to the Marshaler interface
func (irj *indexRecordJSON) fromJSON() (*IndexRecord, error) {
	result := IndexRecord{
		FilePath:     irj.FilePath,
		SizeInBytes:  irj.SizeInBytes,
		Sha1Hash:     [20]byte{},
		ProgressFile: nil,
		ChunkSize:    irj.ChunkSize,
	}

	hash, err := hex.DecodeString(irj.Sha1Hash)
	if err != nil {
		return nil, err
	}
	copy(result.Sha1Hash[:], hash)

	return &result, nil
}

// indexRecordJSON is a private intermediary representation of an IndexRecord for JSON encoding. It
// does not store a pointer to a progress file, since the FilePath is exactly that.
type indexRecordJSON struct {
	FilePath    string
	SizeInBytes int64
	Sha1Hash    string
	ChunkSize   int
}
