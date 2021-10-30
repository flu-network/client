package messages

import (
	"encoding/binary"

	"github.com/flu-network/client/common"
)

type OpenLineRequest struct {
	RequestID uint16
	Sha1Hash  *common.Sha1Hash // which file we want
	Chunk     uint16           // which chunk of that file we want
}

func (r *OpenLineRequest) Serialize() []byte {

}

// Type returns a uint8 that identifies this message type
func (r *OpenLineRequest) Type() byte {
	return openLineRequest
}

// ResponseType returns a uint8 that identidies the type of response expected for this message
func (r *OpenLineRequest) ResponseType() byte {
	return openLineResponse
}

type OpenLineResponse struct {
	RequestID uint16
}

// Type returns a uint8 that identifies this message type
func (r *OpenLineResponse) Type() byte {
	return openLineResponse
}

// Serialize converts its subject into a []byte for transmission over the wire
func (r *OpenLineResponse) Serialize() []byte {
	result := make([]byte, 5)

	// message type
	result[0] = openLineResponse

	// request ID
	binary.BigEndian.PutUint16(result[1:3], r.RequestID)

	// number of entries. Implicit: you cannot store more than ~65k files
	binary.BigEndian.PutUint16(result[3:5], uint16(len(r.Files)))

	for _, entry := range r.Files {
		result = append(result, entry.Serialize()...)
	}

	return result
}
