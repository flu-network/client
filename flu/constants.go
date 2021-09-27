package flu

import "net"

// exports

// UDPPort is the port used by Flu for all UDP communication
const UDPPort = int(61697)

// private

var broadcastAddress = net.UDPAddr{
	IP:   []byte{127, 0, 0, 1},
	Port: UDPPort,
	Zone: "",
}

// message type identifiers
const discoverHostRequest = uint8(0)
const discoverHostResponse = uint8(1)
