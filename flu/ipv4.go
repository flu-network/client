package flu

import (
	"fmt"
	"net"
)

type ipv4 [4]byte

func newIpv4(netIP *net.IP) (ipv4, error) {
	converted := netIP.To4()
	result := [4]byte{}
	if converted == nil {
		return result, fmt.Errorf("provided IP address is not an IPV4 address")
	}

	switch cap(*netIP) {
	case 4:
		copy(result[:], (*netIP)[0:4])
	case 16:
		copy(result[:], (*netIP)[12:16])
	default:
		panic(fmt.Errorf("unknown net.IP %v", netIP))
	}

	return result, nil
}
