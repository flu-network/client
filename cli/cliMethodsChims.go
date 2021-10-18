package cli

import (
	"fmt"
	"strings"

	"github.com/flu-network/client/common"
)

// ChimRequest is a host-discovery message (since chimneys are portals to the floo network in Harry
// Potter). All hosts are expected to respond with their available info. The details of the
// ChimRequest indicate just how detailed we want the response to be.
type ChimRequest struct {
	// if provided, only hosts with info on this file will respond.
	Sha1Hash *common.Sha1Hash
}

// ChimResponse lists the available hosts on the network. If a sha1Hash was provided, hosts will
// include the chunks of that file that they have available.
type ChimResponse struct {
	HostIP   [4]byte
	HostPort uint16
	Chunks   []uint16
}

// Sprintf returns a pretty-printed, user-facing string representation of a ChimResponse
func (c *ChimResponse) Sprintf() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("IP: %d.%d.%d.%d:%d\n",
		c.HostIP[0], c.HostIP[1], c.HostIP[2], c.HostIP[3], c.HostPort))
	for i := 0; i < len(c.Chunks); i += 2 {
		b.WriteString(fmt.Sprintf("  %d:%d\n", c.Chunks[i], c.Chunks[i+1]))
	}
	return b.String()
}

// ChimResponseList is a list of ChimResponses
type ChimResponseList struct {
	Responses []ChimResponse
}

// Sprintf returns a pretty-printed, user-facing string representation of a ChimResponseList
func (c *ChimResponseList) Sprintf() string {
	var b strings.Builder
	for _, resp := range c.Responses {
		b.WriteString(resp.Sprintf())
	}
	return b.String()
}

// Chims lists available hosts on the network. If a sha1 is provided, only hosts that have at least
// some of that file will respond, and their responses will be scoped to that one file.
func (m *Methods) Chims(req *ChimRequest, resp *ChimResponseList) error {
	r := m.fluServer.DiscoverHosts(req.Sha1Hash, []uint16{})
	resp.Responses = make([]ChimResponse, len(r))
	for i, peer := range r {
		resp.Responses[i] = ChimResponse{
			HostIP:   peer.Address,
			HostPort: peer.Port,
			Chunks:   peer.Chunks,
		}
	}

	return nil
}
