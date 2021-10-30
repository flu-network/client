package cli

import (
	"fmt"
	"strings"

	"github.com/flu-network/client/common"
)

// CleanRequest is just a signal to the daemon. It contains no specific information
type CleanRequest struct{}

// CleanResponseItem contains the FilePath and Sha1Hash information about an indexed file
type CleanResponseItem struct {
	FilePath        string
	IndexedSha1Hash [20]byte
	CurrentSha1Hash [20]byte // If null, file is missing. If different, file is corrupted.
}

// Sprintf returns a pretty-printed, user-facing string representation of a CleanResponseItem
func (r *CleanResponseItem) Sprintf() string {
	var result strings.Builder
	indexedHash := (&common.Sha1Hash{}).FromSlice(r.IndexedSha1Hash[:])
	currentHash := (&common.Sha1Hash{}).FromSlice(r.CurrentSha1Hash[:])

	switch {
	case currentHash.IsBlank(): // file is missing
		result.WriteString(fmt.Sprintf("%s\n", r.FilePath))
		result.WriteString("  - File is missing. Removed from index\n")
	case currentHash.Data != indexedHash.Data: // file is corrupt
		result.WriteString(fmt.Sprintf("%s\n", r.FilePath))
		result.WriteString("  - File has changed since it was shared. Removed from Index\n")
		result.WriteString(fmt.Sprintf("    Indexed Hash: %s\n", indexedHash.String()))
		result.WriteString(fmt.Sprintf("    Current Hash: %s\n", currentHash.String()))
	default:
		// file is fine. Print nothing.
	}

	return result.String()
}

// CleanResponse contains basic information about files that had to be removed
type CleanResponse struct {
	Items []CleanResponseItem
}

// Sprintf returns a pretty-printed, user-facing string representation of a CleanResponse
func (r *CleanResponse) Sprintf() string {
	var result strings.Builder
	for _, item := range r.Items {
		result.WriteString(item.Sprintf())
	}
	return result.String()
}

// Clean checks if each file in the index can be safely shared, removes those that can't from the
// index and returns errors for those removed files
func (m *Methods) Clean(req *CleanRequest, resp *CleanResponse) error {
	files, err := m.cat.ListFiles()
	if err != nil {
		return err
	}

	for _, f := range files {
		currentHash, err := m.cat.Rehash(&f)
		if err != nil && f.ProgressFile.Progress.Full() {
			m.cat.UnshareFile(&f)
		}
		resp.Items = append(resp.Items, CleanResponseItem{
			FilePath:        f.FilePath,
			IndexedSha1Hash: f.Sha1Hash.Data,
			CurrentSha1Hash: currentHash.Data,
		})
	}

	return nil
}
