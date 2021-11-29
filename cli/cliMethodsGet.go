package cli

import "github.com/flu-network/client/common"

// GetRequest contains the information necessary to initiate a flu transfer to get a file
type GetRequest struct {
	Sha1Hash *common.Sha1Hash // sha1 hash of the file being downloaded
}

// GetResponse is an empty struct
type GetResponse struct{}

// Sprintf returns a pretty-printed, user-facing string representation of a GetResponse
func (res *GetResponse) Sprintf() string {
	return "Flu transfer initiated\n"
}

// Get initiates a flu transfer for the specified file
func (m *Methods) Get(req *GetRequest, res *GetResponse) error {
	return m.fluServer.StartDownload(req.Sha1Hash)
}
