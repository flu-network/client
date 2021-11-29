package messages

// Message describes the methods that all flu messages have in common.
type Message interface {
	Serialize() []byte // Serialize converts its subject into a []byte for transmission
	Type() uint8       // Type returns a uint8 that identidies the type of message
}

// message type identifiers
const discoverHostRequest = uint8(0)
const discoverHostResponse = uint8(1)
const listFilesRequest = uint8(2)
const listFilesResponse = uint8(3)
const openLineRequest = uint8(4)
const dataPacket = uint8(5)
const dataPacketAck = uint8(6)
