package flu

import (
	"fmt"
	"net"

	"github.com/flu-network/client/flu/messages"
)

func (s *Server) ContinueUpload(
	ack *messages.DataPacketAck,
	conn *net.UDPConn,
	returnAddr *net.UDPAddr,
) error {
	remoteHostIP, err := newIpv4(&returnAddr.IP)
	if err != nil {
		return err
	}

	key := uploadKey{
		remoteHost: remoteHostIP,
		remotePort: uint16(returnAddr.Port),
	}

	sc, ok := s.uploads[key]
	if !ok {
		return fmt.Errorf("no upload corresponds to %v:%d", key.remoteHost, key.remotePort)
	}

	sc.packetChan <- *ack
	return nil
}
