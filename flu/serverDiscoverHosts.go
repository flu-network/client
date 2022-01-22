package flu

import (
	"net"
	"time"

	"github.com/flu-network/client/common"
	"github.com/flu-network/client/flu/messages"
)

// DiscoverHosts broadcasts a DiscoverHostRequest on the local network, collects responses for
// a few seconds, and returns the collected results. Both arguments are optional and serve as
// filters.
func (s *Server) DiscoverHosts(
	hash *common.Sha1Hash,
	chunks []uint16,
) []messages.DiscoverHostResponse {
	// construct a request
	req := messages.DiscoverHostRequest{
		Sha1Hash:  *hash,
		RequestID: s.generateRequestID(),
		Chunks:    chunks,
	}

	// add a response harness for it
	responseChan := s.registerResponseChan(req.RequestID, req.ResponseType())

	// send it into the ether
	var broadcastAddress = net.UDPAddr{
		IP:   []byte{255, 255, 255, 255}, // broadcast IP
		Port: s.port,                     // all hosts should use the same port
	}

	conn, err := net.DialUDP("udp", nil, &broadcastAddress)
	check(err)
	defer conn.Close()
	conn.Write(req.Serialize())

	// set a timeout and wait for the response
	waitChan := time.After(2 * time.Second)
	result := make([]messages.DiscoverHostResponse, 0)

	for {
		select {
		case <-waitChan:
			// if timed out, clean up
			s.unregisterResponseChan(req.RequestID, req.ResponseType())
			return result
		case res := <-responseChan:
			// else cast response into desired type
			parsedResponse := res.(*messages.DiscoverHostResponse)
			result = append(result, *parsedResponse)
		}
	}
}

func (s *Server) RespondToDiscoverHosts(
	req *messages.DiscoverHostRequest,
	returnAddr *net.UDPAddr,
) error {
	ip := s.LocalIP()
	resp := messages.DiscoverHostResponse{
		Address:   [4]byte{(*ip)[0], (*ip)[1], (*ip)[2], (*ip)[3]},
		Port:      uint16(s.port),
		RequestID: req.RequestID,
		Chunks:    []uint16{},
	}

	if !req.Sha1Hash.IsBlank() {
		if ir, err := s.cat.Contains(&req.Sha1Hash); err == nil {
			if len(req.Chunks) > 0 { // if they asked for chunks
				resp.Chunks = ir.Progress.Overlap(req.Chunks) // return overlap
			} else {
				resp.Chunks = ir.Progress.Ranges() // return all ranges
			}
		}
	}

	return s.sendToPeer(returnAddr.IP, resp.Serialize())
}
