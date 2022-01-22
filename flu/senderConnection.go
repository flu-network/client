package flu

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/flu-network/client/common"
	"github.com/flu-network/client/flu/messages"
)

type SenderConnection struct {
	reader     *common.ChunkReader
	windowSize uint16
	windowCap  uint16
	conn       *net.UDPConn
	addr       *net.UDPAddr
	packetChan chan messages.DataPacketAck
	cancelChan chan struct{}
}

func NewSenderConnection(reader *common.ChunkReader,
	windowCap uint16,
	conn *net.UDPConn,
	addr *net.UDPAddr,
) *SenderConnection {
	return &SenderConnection{
		reader:     reader,
		windowSize: 0,
		windowCap:  windowCap,
		conn:       conn,
		addr:       addr,
		packetChan: make(chan messages.DataPacketAck, windowCap),
		cancelChan: make(chan struct{}, 1),
	}
}

// kickstart sends a message to the client telling it about the data to expect so it can set up
// its own connection harnessing. It spawns a 'worker' routine to handle this connection. The main
// routine then sends acks to the worker via the SenderConnection's packetChan.
func (sc *SenderConnection) kickstart(
	hash *common.Sha1Hash,
	size int64,
) error {
	sc.reader.Reset()

	firstPacket := messages.DataPacket{
		Offset: 0,
		Data:   make([]byte, 1024),
	}

	copy(firstPacket.Data[:20], hash.Slice())
	binary.BigEndian.PutUint32(firstPacket.Data[20:24], uint32(size))

	actualDataSpace := firstPacket.Data[24:]
	byteCount, _, err := sc.reader.Read(actualDataSpace)

	if err != nil && err != io.EOF {
		// EOF is fine because the goroutine handling the error will exit when it tries to read
		// again and gets its own EOF with 0 bytes of data
		return err
	}

	firstPacket.Data = firstPacket.Data[:24+byteCount]

	_, err = sc.conn.WriteTo(firstPacket.Serialize(), sc.addr)
	if err != nil {
		return err
	}

	sc.windowSize++

	go func() {
		for {
			select {
			case ack := <-sc.packetChan:
				err := sc.kick(ack)
				if err != nil {
					panic(err) // TODO: Do something with these errors
				}
			case <-sc.cancelChan:
				return
			}
		}
	}()

	return nil
}

func (sc *SenderConnection) terminate() {
	sc.cancelChan <- struct{}{}
}

// kick receives an ack from the client and responds accordingly. Passed by value because acks are
// << 64 bits
func (sc *SenderConnection) kick(ack messages.DataPacketAck) error {
	sc.windowSize--

	packet := messages.DataPacket{Offset: 0, Data: nil}
	packetBuffer := make([]byte, 1024)

	for sc.windowSize < sc.windowCap {

		byteCount, offset, err := sc.reader.Read(packetBuffer)

		if err == io.EOF {
			defer sc.terminate() // make this the last message
		} else if err != nil {
			return err
		}

		packet.Data = packetBuffer[:byteCount] // clip to number of bytes read
		packet.Offset = offset

		_, err = sc.conn.WriteTo(packet.Serialize(), sc.addr)
		if err != nil {
			return err
		}

		sc.windowSize++
	}

	return nil
}
