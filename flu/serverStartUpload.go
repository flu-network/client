package flu

import (
	"fmt"
	"math"
	"net"

	"github.com/flu-network/client/flu/messages"
)

func (s *Server) StartUpload(
	msg *messages.OpenConnectionRequest,
	conn *net.UDPConn,
	returnAddr *net.UDPAddr,
) error {

	ir, err := s.cat.Contains(msg.Sha1Hash)
	if err != nil {
		return err
	}

	reader, err := s.cat.GetChunkReader(&ir.Sha1Hash, int64(msg.Chunk))
	if err != nil {
		return err
	}

	if reader.Size > math.MaxUint32 {
		err = fmt.Errorf("chunk size (%d) exceeds max 32-bit int (%d)", reader.Size, math.MaxInt32)
		return err
	}

	remoteHostIP, err := newIpv4(&returnAddr.IP)
	if err != nil {
		return err
	}

	key := uploadKey{
		remoteHost: remoteHostIP,
		remotePort: uint16(returnAddr.Port),
	}

	if _, ok := s.uploads[key]; !ok {
		s.uploads[key] = NewSenderConnection(reader, msg.WindowCap, conn, returnAddr)
	}

	return s.uploads[key].kickstart(&reader.Hash, int64(reader.Size))
}
