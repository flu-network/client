package cli

// GetRequest contains the information necessary for the daemon to begin a transfer.
type GetRequest struct {
	Sha1Hash [20]byte
}

// GetResponse contains the information necessary for the client to setup the transfer locally. For
// a dummy TCP transfer this will be an empty response, but for a flu transfer this will contain
// everything necessary to set up a local progressFile and indexRecord for the file being requested.
type GetResponse struct {
}

// Get initializes a flu transfer of the file specified in the GetRequest.
func (m *Methods) Get(args *GetRequest, reply *GetResponse) error {
	return nil
}
