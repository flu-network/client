package flu

import (
	"fmt"
	"net"

	"github.com/flu-network/client/catalogue"
)

func HandleMessage(cat *catalogue.Cat, message []byte) ([]byte, error) {
	reader := byteReader{Data: message}
	messageType := reader.readByte()

	switch messageType {
	case discoverHostRequest:
		// hash := reader.readSha1Hash()
		reqID := reader.readUint16()
		// chunks := reader.readSliceUint16()
		// message := DiscoverHostRequest{
		// 	Sha1Hash:  *hash,
		// 	RequestID: reqID,
		// 	Chunks:    chunks,
		// }

		// record, err := cat.Contains(hash)
		// if err != nil {
		// }

		ip := localIP()

		resp := DiscoverHostResponse{
			Address:   [4]byte{(*ip)[0], (*ip)[1], (*ip)[2], (*ip)[3]},
			Port:      uint16(UDPPort),
			RequestID: reqID,
		}

		return resp.Serialize(), nil

	case discoverHostResponse:
		add := reader.readBytes(4)
		port := reader.readUint16()
		reqID := reader.readUint16()
		message := DiscoverHostResponse{
			Address:   [4]byte{add[0], add[1], add[2], add[3]},
			Port:      port,
			RequestID: reqID,
		}
		fmt.Println("Got response", message)

	default:
		return nil, fmt.Errorf("Message of unknown type discarded: %d", messageType)
	}

	return nil, nil
}

func localIP() *net.IP {
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
