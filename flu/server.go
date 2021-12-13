package flu

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/flu-network/client/catalogue"
	"github.com/flu-network/client/common"
	"github.com/flu-network/client/flu/messages"
)

const maxRequestID = int(1 << 16)

// Server dispatches and responds to flu messages. It contains enough internal state to map an
// incoming UDP message to a previous outgoing message, thereby providing a request/response
// paradigm when needed. Every request/response is tagged with a monotonically-increasing requestID
// so many conversations can occur concurrently.
// All incoming UDP messages should be sent to Server for handling.
type Server struct {
	port int
	cat  *catalogue.Cat

	// reqID and reqIDLock are responsible for generating locally-unique request IDs for incoming
	// outgoing request-response communications between peers
	reqID     int
	reqIDLock sync.Mutex

	// resMap and resMapLock are responsible for mapping outgoing requests to incoming responses.
	// They map an outgoing requestID and response type to an incoming message and deliver the
	// response to the Message chan stored in the map.
	resMap     map[requestKey]chan messages.Message
	resMapLock sync.Mutex

	// transfer lock, downloads and uploads ensures safe access to the download and upload maps.
	// These maps are used to keep track on ongoing transfers, irrespective of their state.
	transferLock sync.Mutex
	downloads    map[downloadKey]struct{} // corresponds to a single chunk from a single host
	uploads      map[uploadKey]*SenderConnection
}

// requestKey is used to uniquely identify a request that is awaiting one or more responses in a
// simplified request/response paradigm. E.g., host discovery, or one-off peer messages.
type requestKey struct {
	reqID       uint16
	messageType uint8
}

type downloadKey struct {
	hash       common.Sha1Hash
	remoteHost ipv4
}

type uploadKey struct {
	remoteHost ipv4
	remotePort uint16
}

// NewServer returns a *Server
func NewServer(port int, cat *catalogue.Cat) *Server {
	return &Server{
		port:         port,
		cat:          cat,
		reqID:        0,
		reqIDLock:    sync.Mutex{},
		resMap:       make(map[requestKey](chan messages.Message)),
		resMapLock:   sync.Mutex{},
		transferLock: sync.Mutex{},
		downloads:    make(map[downloadKey]struct{}),
		uploads:      make(map[uploadKey]*SenderConnection),
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

// deliverResponse delivers a response message to the goRoutine that originally sent the request
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

func (s *Server) sendToPeer(ip net.IP, message []byte) error {
	sock, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: ip, Port: s.port})
	if err != nil {
		return err
	}
	defer sock.Close()
	_, err = sock.Write(message)
	if err != nil {
		return err
	}
	return nil
}

// Port retuns the port the server is running on
func (s *Server) Port() int {
	return s.port
}

// HandleMessage does exactly what it says. It expects parameters to be passed by value because
// it is assumed it will be run concurrently.
func (s *Server) HandleMessage(message []byte, conn *net.UDPConn, returnAddr *net.UDPAddr) error {
	parsedMessage, err := messages.Parse(message)
	if err != nil {
		return err
	}

	switch msg := parsedMessage.(type) {
	case *messages.DiscoverHostRequest:
		return s.RespondToDiscoverHosts(msg, returnAddr)
	case *messages.ListFilesRequest:
		return s.RespondToListFilesOnHost(msg, conn, returnAddr)
	case *messages.OpenConnectionRequest:
		return s.StartUpload(msg, conn, returnAddr)
	case *messages.DataPacketAck:
		return s.ContinueUpload(msg, conn, returnAddr)
	case *messages.DiscoverHostResponse:
		return s.deliverResponse(msg.RequestID, parsedMessage)
	case *messages.ListFilesResponse:
		return s.deliverResponse(msg.RequestID, parsedMessage)

	// freak out if we don't know how to handle a parsed response
	default:
		panic("Rohan messed up! We parsed a message but couldn't handle it.")
	}
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

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
