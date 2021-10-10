package flu

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/flu-network/client/catalogue"
	"github.com/flu-network/client/common"
	"github.com/flu-network/client/flu/messages"
)

const maxRequestID = int(1 << 16)

// Server dispatches and responds to flu messages. It contains enough internal state to map an
// incoming UDP message to a previous outgoing message, thereby providing a request/response
// paradigm when needed. Every request/response is tagged with a monotonically-increasing requestID
// so many conversations can occur concurrently.
// All incoming UDP messages should be sent to Server for handling, and all outdoing UDP messages
// sent from here.
type Server struct {
	port int
	cat  *catalogue.Cat

	reqID     int
	reqIDLock sync.Mutex

	resMap     map[requestKey]chan messages.Message
	resMapLock sync.Mutex
}

type requestKey struct {
	reqID       uint16
	messageType uint8
}

// NewServer returns a *Server
func NewServer(port int, cat *catalogue.Cat) *Server {
	return &Server{
		port:       port,
		cat:        cat,
		reqID:      0,
		reqIDLock:  sync.Mutex{},
		resMap:     make(map[requestKey](chan messages.Message)),
		resMapLock: sync.Mutex{},
	}
}

func (s *Server) generateRequestID() uint16 {
	s.reqIDLock.Lock()
	defer s.reqIDLock.Unlock()
	result := s.reqID
	s.reqID = (s.reqID + 1) % maxRequestID
	return uint16(result)
}

func (s *Server) registerResponseChan(reqID uint16, msgType uint8) chan messages.Message {
	key := requestKey{reqID: reqID, messageType: msgType}
	s.resMapLock.Lock()
	defer s.resMapLock.Unlock()
	responseChan := make(chan messages.Message, 1)
	s.resMap[key] = responseChan
	return responseChan
}

func (s *Server) unregisterResponseChan(reqID uint16, msgType uint8) {
	key := requestKey{reqID: reqID, messageType: msgType}
	s.resMapLock.Lock()
	defer s.resMapLock.Unlock()
	delete(s.resMap, key)
}

func (s *Server) deliverResponse(reqID uint16, msg messages.Message) error {
	key := requestKey{reqID: reqID, messageType: msg.Type()}
	s.resMapLock.Lock()
	defer s.resMapLock.Unlock()
	if responseChan, ok := s.resMap[key]; ok {
		responseChan <- msg
		return nil
	}
	return fmt.Errorf("ResponseChan {%d:%d} expired", reqID, msg)
}

// HandleMessage does exactly what it says. It expects parameters to be passed by value because
// it is assumed it will be run concurrently.
func (s *Server) HandleMessage(message []byte, returnIP net.IP) error {
	parsedMessage, err := messages.Parse(message)
	if err != nil {
		return err
	}

	var reply []byte = nil

	switch msg := parsedMessage.(type) {
	// handle requests
	case *messages.DiscoverHostRequest:
		ip := s.LocalIP()
		resp := messages.DiscoverHostResponse{
			Address:   [4]byte{(*ip)[0], (*ip)[1], (*ip)[2], (*ip)[3]},
			Port:      uint16(s.port),
			RequestID: msg.RequestID,
		}
		reply = resp.Serialize()

	// deliver responses
	case *messages.DiscoverHostResponse:
		return s.deliverResponse(msg.RequestID, parsedMessage)

	// freak out if we don't know how to handle a parsed response
	default:
		panic("Rohan messed up! We parsed a message but couldn't handle it.")
	}

	// only reply if there's something to send
	if reply != nil {
		returnAddr := net.UDPAddr{IP: returnIP, Port: s.port, Zone: ""}
		respSock, err := net.DialUDP("udp", nil, &returnAddr)
		if err != nil {
			return err
		}
		defer respSock.Close()
		respSock.Write(reply)
	}

	return nil
}

// LocalIP returns the IPV4 address of the running process (if available). This is not the loopback
// address (192.168.0.1; 'localhost'), but the address assigned to this host by the LAN's router.
func (s *Server) LocalIP() *net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			result := ipnet.IP.To4()
			if result != nil {
				return &result
			}
		}
	}

	return nil
}

// FindAvailableHosts broadcasts a DiscoverHostRequest on the local network, collects responses for
// one second, and returns the collected results.
// TODO: filter results down to just the specified hash.
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
	waitChan := time.After(1 * time.Second)
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

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
