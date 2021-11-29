package messages

import (
	"encoding/binary"

	"github.com/flu-network/client/common"
)

// ListFilesRequest is sent to a single host requesting them to list the files they have available
// to download. Returned file may be partially or completely available.
type ListFilesRequest struct {
	// The requestID is only used by the client to tie a response to an outgoing request
	RequestID uint16
	Sha1Hash  *common.Sha1Hash
}

// Serialize converts its subject into a []byte for transmission over the wire
func (r *ListFilesRequest) Serialize() []byte {
	result := make([]byte, 23)

	// message type
	result[0] = listFilesRequest

	// request ID
	binary.BigEndian.PutUint16(result[1:3], r.RequestID)

	// sha1 hash
	copy(result[3:23], r.Sha1Hash.Data[:])

	return result
}

// Type returns a uint8 that identifies this message type
func (r *ListFilesRequest) Type() byte {
	return listFilesRequest
}

// ResponseType returns a uint8 that identidies the type of response expected for this message
func (r *ListFilesRequest) ResponseType() byte {
	return listFilesResponse
}

// ListFilesEntry shows the 'safe-to-share' info about a given file in the index
type ListFilesEntry struct {
	SizeInBytes      uint64
	ChunkCount       uint32
	ChunkSizeInBytes uint32 // ChunkCount * ChunkSizeInBytes == SizeInBytes
	// The number of chunks of the file that are downloaded and available for sharing
	ChunksDownloaded uint32
	Sha1Hash         *common.Sha1Hash
	FileName         string
}

// Serialize converts its subject into a []byte for transmission over the wire
func (lfe *ListFilesEntry) Serialize() []byte {
	name := SerializeString255(lfe.FileName)
	result := make([]byte, 40, len(name)+40)

	binary.BigEndian.PutUint64(result[0:8], lfe.SizeInBytes)
	binary.BigEndian.PutUint32(result[8:12], lfe.ChunkCount)
	binary.BigEndian.PutUint32(result[12:16], lfe.ChunkSizeInBytes)
	binary.BigEndian.PutUint32(result[16:20], lfe.ChunksDownloaded)
	copy(result[20:40], lfe.Sha1Hash.Array()[:])
	result = append(result, name...)

	return result
}

// ListFilesResponse contains a list of files availble in the index
type ListFilesResponse struct {
	RequestID uint16
	Files     []ListFilesEntry
}

// Type returns a uint8 that identifies this message type
func (r *ListFilesResponse) Type() byte {
	return listFilesResponse
}

// Serialize converts its subject into a []byte for transmission over the wire
func (r *ListFilesResponse) Serialize() []byte {
	result := make([]byte, 5)

	// message type
	result[0] = listFilesResponse

	// request ID
	binary.BigEndian.PutUint16(result[1:3], r.RequestID)

	// number of entries. Implicit: you cannot store more than ~65k files
	binary.BigEndian.PutUint16(result[3:5], uint16(len(r.Files)))

	for _, entry := range r.Files {
		result = append(result, entry.Serialize()...)
	}

	return result
}
