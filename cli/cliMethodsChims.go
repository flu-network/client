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
	Sha1Hash [20]byte
}

// ChimResponse lists the available files for a host on the network. If a sha1Hash was provided,
// the responses will be scoped to that one file.
type ChimResponse struct {
	HostIP   [4]byte
	HostPort uint16
	Items    []ListItem
}

// Sprintf returns a pretty-printed, user-facing string representation of a ChimResponse
func (c *ChimResponse) Sprintf() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("IP: %d.%d.%d.%d:%d\n",
		c.HostIP[0], c.HostIP[1], c.HostIP[2], c.HostIP[3], c.HostPort))
	for _, item := range c.Items {
		b.WriteString(item.Sprintf())
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
// some of that file will respond, and their responses will be scoped to that one file. The daemon
// running locally will not respond to this request. For that, use `flu list`.
func (m *Methods) Chims(req *ChimRequest, resp *ChimResponseList) error {
	// TODO: Reach out to the daemon to actually send and fulfill the request
	r := m.fluServer.FindAvailableHosts(&common.Sha1Hash{}, []uint16{})
	resp.Responses = make([]ChimResponse, len(r))
	for i, peer := range r {
		resp.Responses[i] = ChimResponse{
			HostIP:   peer.Address,
			HostPort: peer.Port,
			Items:    []ListItem{},
		}
	}

	return nil
}
