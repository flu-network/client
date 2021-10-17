package flu

import (
	"net"
	"time"

	"github.com/flu-network/client/common"
	"github.com/flu-network/client/flu/messages"
)

// FindAvailableHosts broadcasts a DiscoverHostRequest on the local network, collects responses for
// one second, and returns the collected results.
func (s *Server) FindAvailableHosts(
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
