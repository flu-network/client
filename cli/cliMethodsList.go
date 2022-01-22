package cli

import (
	"net"
	"strings"

	"github.com/flu-network/client/common"
)

// ListRequest contains the information necessary for the daemon to find, hash, index and List the
// file pointed to by FilePath.
type ListRequest struct {
	IP       net.IP
	Sha1Hash *common.Sha1Hash
}

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
	if req.IP == nil {
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
				ChunkCount:       rec.Progress.Size(),
				ChunkSizeInBytes: rec.ChunkSize,
				ChunksDownloaded: rec.Progress.Count(),
			}
		}
	} else {
		addr := req.IP.To4()
		ip := [4]byte{addr[0], addr[1], addr[2], addr[3]}
		r, err := m.fluServer.ListFilesOnHost(ip, uint16(m.fluServer.Port()), req.Sha1Hash)
		if err != nil {
			return err
		}

		resp.Items = make([]ListItem, len(r.Files))
		for i, file := range r.Files {
			resp.Items[i] = ListItem{
				FilePath:         file.FileName,
				SizeInBytes:      int64(file.SizeInBytes),
				Sha1Hash:         *file.Sha1Hash.Array(),
				ChunkCount:       int(file.ChunkCount),
				ChunkSizeInBytes: int(file.ChunkSizeInBytes),
				ChunksDownloaded: int(file.ChunksDownloaded),
			}
		}
	}

	return nil

}
