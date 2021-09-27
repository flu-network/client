package flu

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/flu-network/client/catalogue"
	"github.com/flu-network/client/common"
)

const discoverHostTimeout = time.Second // Allow hosts one second to respond

// DiscoverHostRequest is broadcast to all hosts on the LAN to ask participating hosts which chunks
// in the given range they have of the specified file hash.
type DiscoverHostRequest struct {
	// The sha1 hash of the file we're interested in. If the hash is FFF... then hosts are requested
	// to respond with information about all files they have available
	Sha1Hash common.Sha1Hash

	// The requestID is only used by the client to tie a response to an outgoing request
	RequestID uint16

	// The ranges of chunks we're interested in. If no chunks are specified then hosts are requested
	// to return the ranges of all chunks that they have
	Chunks []uint16
}

// Serialize converts its subject into a []byte for transmission over the wire
func (r *DiscoverHostRequest) Serialize() []byte {
	result := make([]byte, 24)

	// message type
	result[0] = discoverHostRequest

	// sha1 hash
	copy(result[1:21], r.Sha1Hash.Slice())

	// request ID
	binary.BigEndian.PutUint16(result[21:23], r.RequestID)

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

// DiscoverHostResponse is returned in response to a DiscoverHostRequest
type DiscoverHostResponse struct {
	Address   [4]byte
	Port      uint16
	RequestID uint16
}

// Serialize converts its subject into a []byte for transmission over the wire
func (r *DiscoverHostResponse) Serialize() []byte {
	result := make([]byte, 9)

	// message type
	result[0] = discoverHostResponse

	// address
	copy(result[1:5], r.Address[:])

	// port
	binary.BigEndian.PutUint16(result[5:7], r.Port)

	// request ID
	binary.BigEndian.PutUint16(result[7:9], r.RequestID)

	return result
}

func Listen(cat *catalogue.Cat) {
	addr := net.UDPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: UDPPort,
		Zone: "",
	}
	c1, err := net.ListenUDP("udp", &addr)
	defer c1.Close()
	check(err)

	buffer := make([]byte, 1024)
	_, returnAddress, err := c1.ReadFromUDP(buffer)
	check(err)

	resp, err := HandleMessage(cat, buffer)
	check(err)

	c2, err := net.DialUDP("udp", nil, returnAddress)
	defer c2.Close()
	check(err)

	c2.Write(resp)

	fmt.Printf("%v\n", returnAddress)
}

func Send() {
	conn, err := net.DialUDP("udp", nil, &broadcastAddress)
	check(err)
	defer conn.Close()

	req := DiscoverHostRequest{
		Sha1Hash:  *(&common.Sha1Hash{}).FromString("b32d902059d1dff19eedffa39c561ef4f0ddfc29"),
		RequestID: 123,
		Chunks:    []uint16{},
	}

	bytes := req.Serialize()

	conn.Write(bytes)

}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
