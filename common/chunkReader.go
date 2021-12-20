package common

import (
	"crypto/sha1"
	"fmt"
	"io"
	"math"
)

// ChunkReader is a thin wrapper around an io.SectionReader. For convenience, it also contains the
// sha1 Hash of the bytes underlying the io.SectionReader, and also the number of bytes it can read
type ChunkReader struct {
	Reader io.SectionReader
	Hash   Sha1Hash
	Size   int64
}

func (cr *ChunkReader) Offset() uint32 {
	offset, err := cr.Reader.Seek(0, io.SeekCurrent)

	if err != nil {
		panic(err)
	}
	if offset > math.MaxUint32 {
		panic(fmt.Errorf(
			"ChunkReader cannot yield %d, greater than max uint32 %d",
			offset,
			math.MaxUint32,
		))
	}

	return uint32(offset)
}

func (cr *ChunkReader) Read(buffer []byte) (int, uint32, error) {
	offset := cr.Offset()
	bytesRead, err := cr.Reader.Read(buffer)
	return bytesRead, offset, err
}

func (cr *ChunkReader) Reset() error {
	_, err := cr.Reader.Seek(0, 0)

	if err != nil {
		return err
	}

	if cr.Offset() != 0 {
		return fmt.Errorf("internal error: resetting ChunkReader had no effect")
	}

	return nil
}

func NewChunkReader(reader *io.SectionReader) *ChunkReader {
	hash := sha1.New()
	size := 0

	hashBuffer := make([]byte, 4096) // A reasonably-recent mac's block size

	for {
		bytesRead, err := reader.Read(hashBuffer)
		if bytesRead > 0 {
			size += bytesRead
			wrote, err := hash.Write(hashBuffer[:bytesRead])
			switch {
			case err != nil && err != io.EOF:
				panic(err)
			case wrote != bytesRead:
				panic(fmt.Errorf("internal error writing to hash"))
			default:
				// do nothing
			}
		}

		if err == io.EOF {
			break
		}
	}

	reader.Seek(0, 0)
	finalHash := (&Sha1Hash{}).FromSlice(hash.Sum(nil))

	return &ChunkReader{
		Reader: *reader,
		Hash:   *finalHash,
		Size:   int64(size),
	}
}
