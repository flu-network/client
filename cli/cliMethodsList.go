package cli

import "strings"

// ListRequest contains the information necessary for the daemon to find, hash, index and List the
// file pointed to by FilePath.
type ListRequest struct{}

// ListResponse is a slice of ListItems, each of which details a file that is shared by the daemon.
// The contents of files in the ListResponse are guaranteed unique. Not all files that have been
// listed have been completely downloaded.
type ListResponse struct {
	Items []ListItem
}

// Sprintf returns a pretty-printed, user-facing string representation of a ListResponse
func (lr *ListResponse) Sprintf() string {
	var b strings.Builder
	for _, item := range lr.Items {
		b.WriteString(item.Sprintf())
	}
	return b.String()
}

// List lists the files that have been indexed by the daemon. Not all indexed files have been
// downloaded in their entirety.
func (m *Methods) List(req *ListRequest, resp *ListResponse) error {
	records, err := m.cat.ListFiles()
	if err != nil {
		return err
	}

	resp.Items = make([]ListItem, len(records))
	for i, rec := range records {

		resp.Items[i] = ListItem{
			FilePath:         rec.FilePath,
			SizeInBytes:      rec.SizeInBytes,
			Sha1Hash:         *rec.Sha1Hash.Array(),
			ChunkCount:       rec.ProgressFile.Progress.Size(),
			ChunkSizeInBytes: rec.ChunkSize,
			ChunksDownloaded: rec.ProgressFile.Progress.Count(),
		}
	}
	return nil
}
