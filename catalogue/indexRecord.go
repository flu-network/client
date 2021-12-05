package catalogue

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/flu-network/client/common"
)

const defaultChunkSize = int(1 << 22) // 4mb in bytes

// indexRecord describes a file that is 'known' by the flu client. The existence of an indexRecord
// does not imply that the file exists locally. To find out which chunks of the file are
// downloaded, consult the progressFile. By convention, the progressFile is always named after
// the sha1 of the completely-downloaded file.
// All methods assume the caller has acquired a mutex granting exclusive access.
type indexRecord struct {
	FilePath     string
	SizeInBytes  int64
	Sha1Hash     common.Sha1Hash
	ProgressFile *progressFile
	ChunkSize    int
}

// IndexRecordExport is a copy of an underlying indexRecord intended for read-only access.
// Manipulations should be performed via an appropriate catalogue method.
type IndexRecordExport struct {
	FilePath     string
	SizeInBytes  int64
	Sha1Hash     common.Sha1Hash
	ProgressFile progressFile
	ChunkSize    int
}

// export returns an IndexRecordExport, which is safe for consumption outside of the catalogue
func (ir *indexRecord) export() *IndexRecordExport {
	return &IndexRecordExport{
		FilePath:     ir.FilePath,
		SizeInBytes:  ir.SizeInBytes,
		Sha1Hash:     ir.Sha1Hash,
		ProgressFile: *ir.ProgressFile,
		ChunkSize:    ir.ChunkSize,
	}
}

// getChunkReader returns a ChunkReader. It should be called via the catalogue so we know it is
// done safely. It is the caller's responsibility to ensure the ChunkReader is eventually closed.
func (ir *indexRecord) getChunkReader(chunk int64) (*common.ChunkReader, error) {
	if !ir.ProgressFile.progress.Get(uint64(chunk)) {
		return nil, fmt.Errorf("missing requested chunk %d", chunk)
	}

	fd, err := os.Open(ir.FilePath)
	if err != nil {
		return nil, err
	}

	start := chunk * int64(ir.ChunkSize)
	secReader, err := io.NewSectionReader(fd, start, int64(ir.ChunkSize)), nil
	if err != nil {
		return nil, err
	}

	return common.NewChunkReader(secReader), nil
}

func generateIndexRecordForFile(path string) (*indexRecord, error) {
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

	return &indexRecord{
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
func hashFile(path string) (*common.Sha1Hash, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	s, h := bufio.NewScanner(f), sha1.New()
	for s.Scan() {
		h.Write(s.Bytes())
	}

	return (&common.Sha1Hash{}).FromSlice(h.Sum(nil)), nil
}

// toJSON returns an indexRecordJSON, which can natively be marshalled into JSON
func (ir *indexRecord) toJSON() *indexRecordJSON {
	return &indexRecordJSON{
		FilePath:    ir.FilePath,
		SizeInBytes: ir.SizeInBytes,
		Sha1Hash:    ir.Sha1Hash.String(),
		ChunkSize:   ir.ChunkSize,
	}
}

// UnmarshalJSON conforms to the Marshaler interface
func (irj *indexRecordJSON) fromJSON() (*indexRecord, error) {
	result := indexRecord{
		FilePath:     irj.FilePath,
		SizeInBytes:  irj.SizeInBytes,
		Sha1Hash:     common.Sha1Hash{},
		ProgressFile: nil,
		ChunkSize:    irj.ChunkSize,
	}

	err := result.Sha1Hash.FromStringSafe(irj.Sha1Hash)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// indexRecordJSON is a private intermediary representation of an indexRecord for JSON encoding. It
// does not store a pointer to a progress file, since the FilePath is exactly that.
type indexRecordJSON struct {
	FilePath    string
	SizeInBytes int64
	Sha1Hash    string
	ChunkSize   int
}
