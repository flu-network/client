package flu

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/flu-network/client/common"
	"github.com/flu-network/client/flu/messages"
)

type RecvConnection struct {
	conn          *net.UDPConn
	hash          *common.Sha1Hash
	bytesReceived int
	buffer        []byte
	windowCap     int
	outChan       chan *messages.DataPacket
}

func (r *RecvConnection) Ack(offset uint32) {
	ack := messages.DataPacketAck{
		Offset: offset,
	}
	r.conn.Write(ack.Serialize())
}

func (r *RecvConnection) Read() (*messages.DataPacket, bool) {
	result := <-r.outChan
	if result == nil {
		return result, false
	}
	return result, true
}

func DialPeer(ip [4]byte, port uint16, hash *common.Sha1Hash, chunk uint16) (*RecvConnection, error) {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: ip[:], Port: int(port)})
	if err != nil {
		return nil, err
	}

	result := RecvConnection{
		conn:          conn,
		hash:          nil,
		bytesReceived: 0,
		buffer:        nil,
		windowCap:     10,
		outChan:       make(chan *messages.DataPacket, 10),
	}

	kickstartMsg := messages.OpenConnectionRequest{Sha1Hash: hash, Chunk: chunk, WindowCap: 10}
	result.conn.Write(kickstartMsg.Serialize())

	go func() {
		for {
			buffer := make([]byte, 1029) // NOT 1024: the serialization overhead is 5 bytes
			result.conn.SetReadDeadline(time.Now().Add(time.Second * 5))
			n, _, err := result.conn.ReadFromUDP(buffer)
			if err != nil {
				result.conn.Close()
				result.outChan <- nil
				fmt.Printf("Connection closed: %v-%d:%v\n", hash, chunk, err)
				break
			} else {
				result.outChan <- messages.ParseAsDataPacket(buffer[:n])
			}
		}
	}()

	return &result, nil
}

type SenderConnection struct {
	reader     *common.ChunkReader
	windowSize uint8
	windowCap  uint8
	conn       *net.UDPConn
	addr       *net.UDPAddr
	packetChan chan messages.DataPacketAck
	cancelChan chan struct{}
}

func NewSenderConnection(reader *common.ChunkReader,
	windowCap uint8,
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

	fmt.Printf("Beginning upload of data %v\n", hash)

	copy(firstPacket.Data[:20], hash.Slice())
	binary.BigEndian.PutUint32(firstPacket.Data[20:24], uint32(size))

	actualDataSpace := firstPacket.Data[24:]
	byteCount, _, err := sc.reader.Read(actualDataSpace)

	if err != nil {
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
				sc.kick(&ack) // TODO: Do something with these errors
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

// kick receives an ack from the client and responds accordingly
func (sc *SenderConnection) kick(ack *messages.DataPacketAck) error {
	sc.windowSize--

	for sc.windowSize < sc.windowCap {
		packet := messages.DataPacket{Offset: 0, Data: make([]byte, 1024)}

		byteCount, offset, err := sc.reader.Read(packet.Data)

		if err == io.EOF {
			defer sc.terminate() // make this the last message
		} else if err != nil {
			return err
		}

		packet.Data = packet.Data[:byteCount] // clip to number of bytes read
		packet.Offset = offset

		_, err = sc.conn.WriteTo(packet.Serialize(), sc.addr)
		if err != nil {
			return err
		}

		sc.windowSize++
	}

	return nil
}
