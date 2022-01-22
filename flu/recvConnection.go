package flu

import (
	"fmt"
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
		windowCap:     1024,
		outChan:       make(chan *messages.DataPacket, 10),
	}

	kickstartMsg := messages.OpenConnectionRequest{Sha1Hash: hash, Chunk: chunk, WindowCap: 1024}
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
