package cli

// ShareRequest contains the information necessary for the daemon to find, hash, index and share the
// file pointed to by FilePath.
type ShareRequest struct {
	Filepath string
}

// ShareResponse contains information about the file that was just shared.
type ShareResponse struct {
	FilePath    string
	SizeInBytes int64
	Sha1Hash    [20]byte
}

// Share attempts to add the specified file to flu's index, making it available to anyone on your
// local area network. If an identical file has already been added (even if has a different name),
// flu will refuse to index it twice and return an error telling you which file it was. Sharing a
// file assumes that the file has been downloaded in its entirety, and that its contents will not
// change... ever.
func (m *Methods) Share(req *ShareRequest, resp *ShareResponse) error {
	record, err := m.cat.ShareFile(req.Filepath)
	if err != nil {
		return err
	}
	resp.FilePath = record.FilePath
	resp.SizeInBytes = record.SizeInBytes
	resp.Sha1Hash = record.Sha1Hash
	return nil
}
