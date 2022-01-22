package catalogue

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/flu-network/client/common"
	"github.com/flu-network/client/common/bitset"
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
// Manipulations should be performed via an appropriate catalogue method. Note that since this is
// a copy, once it is created, there is no guarantee that its state is an accurate reflection of
// the underlying object it represents. For strong guarantees of consistency, use the appropriate
// method on the catalogue.
type IndexRecordExport struct {
	FilePath    string
	SizeInBytes int64
	Sha1Hash    common.Sha1Hash
	Progress    bitset.Bitset
	ChunkSize   int
}

// export returns an IndexRecordExport, which is safe for consumption outside of the catalogue
func (ir *indexRecord) export() *IndexRecordExport {
	return &IndexRecordExport{
		FilePath:    ir.FilePath,
		SizeInBytes: ir.SizeInBytes,
		Sha1Hash:    ir.Sha1Hash,
		Progress:    *ir.ProgressFile.Export(),
		ChunkSize:   ir.ChunkSize,
	}
}

func (ir *indexRecord) saveChunk(chunk int64, data []byte) error {
	fd, err := os.OpenFile(ir.FilePath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer fd.Close()
	start := chunk * int64(ir.ChunkSize)
	wrote, err := fd.WriteAt(data, start)
	if err != nil {
		return err
	}
	if wrote != len(data) {
		return fmt.Errorf("wrote only %d of %d bytes", wrote, len(data))
	}
	return nil
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

	hash, err := common.HashFile(cleanPath)
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
