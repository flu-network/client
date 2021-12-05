package cli

import (
	"encoding/hex"
	"fmt"
	"path"
	"strings"
)

// ShareRequest contains the information necessary for the daemon to find, hash, index and share the
// file pointed to by FilePath.
type ShareRequest struct {
	Filepath string
}

// ListItem contains basic information about the file that has been shared.
type ListItem struct {
	FilePath         string
	SizeInBytes      int64
	Sha1Hash         [20]byte
	ChunkCount       int
	ChunkSizeInBytes int // ChunkCount * ChunkSizeInBytes == SizeInBytes
	// The number of chunks of the file that are downloaded and available for sharing
	ChunksDownloaded int
}

// Sprintf returns a pretty-printed, user-facing string representation of a ListItem
func (li *ListItem) Sprintf() string {
	_, fileName := path.Split(li.FilePath)
	output := []string{
		fmt.Sprintf("%s\n", fileName),
		fmt.Sprintf("	Path: %s\n", li.FilePath),
		fmt.Sprintf("	Size (bytes): %d\n", li.SizeInBytes),
		fmt.Sprintf("	Sha1 Hash: %s\n", hex.EncodeToString(li.Sha1Hash[:])),
		fmt.Sprintf("	Chunk Count: %d\n", li.ChunkCount),
		fmt.Sprintf("	Chunks Downloaded: %d\n", li.ChunksDownloaded),
		fmt.Sprintf("	Chunk Size: %d\n", li.ChunkSizeInBytes),
		fmt.Sprintf("	Integrity: %d%%\n", (li.ChunksDownloaded*100/li.ChunkCount*100)/100),
	}

	var b strings.Builder
	for _, line := range output {
		b.WriteString(line)
	}
	return b.String()
}

// Share attempts to add the specified file to flu's index, making it available to anyone on your
// local area network. If an identical file has already been added (even if has a different name),
// flu will refuse to index it twice and return an error telling you which file it was. Sharing a
// file assumes that the file has been downloaded in its entirety, and that its contents will not
// change... ever.
func (m *Methods) Share(req *ShareRequest, resp *ListItem) error {
	record, err := m.cat.ShareFile(req.Filepath)
	if err != nil {
		return err
	}
	resp.FilePath = record.FilePath
	resp.SizeInBytes = record.SizeInBytes
	resp.Sha1Hash = *record.Sha1Hash.Array()
	resp.ChunkCount = record.ProgressFile.Size()
	resp.ChunkSizeInBytes = record.ChunkSize
	resp.ChunksDownloaded = record.ProgressFile.Count()
	return nil
}
