package messages

import (
	"encoding/binary"

	"github.com/flu-network/client/common"
)

// DiscoverHostRequest is broadcast to all hosts on the LAN to ask participating hosts which chunks
// in the given range they have of the specified file hash.
type DiscoverHostRequest struct {
	// The requestID is only used by the client to tie a response to an outgoing request
	RequestID uint16

	// The sha1 hash of the file we're interested in. If the hash is FFF... then hosts are requested
	// to respond with information about all files they have available
	Sha1Hash common.Sha1Hash

	// The ranges of chunks we're interested in. If no chunks are specified then hosts are requested
	// to return the ranges of all chunks that they have
	Chunks []uint16
}

// Serialize converts its subject into a []byte for transmission over the wire
func (r *DiscoverHostRequest) Serialize() []byte {
	result := make([]byte, 24)

	// message type
	result[0] = discoverHostRequest

	// request ID
	binary.BigEndian.PutUint16(result[1:3], r.RequestID)

	// sha1 hash
	copy(result[3:23], r.Sha1Hash.Slice())

	// chunk count
	result[23] = uint8(len(r.Chunks))

	// chunks
	chunks := make([]byte, len(r.Chunks)*2)
	for i, chunkVal := range r.Chunks {
		index := i * 2
		binary.BigEndian.PutUint16(chunks[index:index+2], chunkVal)
	}
	result = append(result, chunks...)

	return result
}

// Type returns a uint8 that identifies this message type
func (r *DiscoverHostRequest) Type() byte {
	return discoverHostRequest
}

// ResponseType returns a uint8 that identidies the type of response expected for this message
func (r *DiscoverHostRequest) ResponseType() byte {
	return discoverHostResponse
}

// DiscoverHostResponse is returned in response to a DiscoverHostRequest
type DiscoverHostResponse struct {
	Address   [4]byte
	Port      uint16
	RequestID uint16
	Chunks    []uint16 // Chunks are only returned if a file is specified in the request
}

// Serialize converts its subject into a []byte for transmission over the wire
func (r *DiscoverHostResponse) Serialize() []byte {
	result := make([]byte, 10)

	// message type
	result[0] = discoverHostResponse

	// request ID
	binary.BigEndian.PutUint16(result[1:3], r.RequestID)

	// address
	copy(result[3:7], r.Address[:])

	// port
	binary.BigEndian.PutUint16(result[7:9], r.Port)

	// chunk count
	result[9] = uint8(len(r.Chunks))

	// chunks
	chunks := make([]byte, len(r.Chunks)*2)
	for i, chunkVal := range r.Chunks {
		index := i * 2
		binary.BigEndian.PutUint16(chunks[index:index+2], chunkVal)
	}
	result = append(result, chunks...)

	return result
}

// Type returns a uint8 that identifies this message type
func (r *DiscoverHostResponse) Type() byte {
	return discoverHostResponse
}
