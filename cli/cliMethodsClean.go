package cli

import (
	"fmt"
	"strings"
)

// CleanRequest is just a signal to the daemon. It contains no specific information
type CleanRequest struct{}

// CleanResponseItem contains the FilePath and Sha1Hash information about an indexed file
type CleanResponseItem struct {
	FilePath        string
	IndexedSha1Hash [20]byte
	CurrentSha1Hash [20]byte // If null, file is missing. If different, file is corrupted.
	ActionTaken     string
}

// Sprintf returns a pretty-printed, user-facing string representation of a CleanResponseItem
func (r *CleanResponseItem) Sprintf() string {
	return r.ActionTaken
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
		var execErr error // set non-nill if something prevented us from cleaning properly
		currentHash, err := m.cat.Rehash(&f.Sha1Hash)

		var actionTaken strings.Builder
		actionTaken.WriteString(fmt.Sprintf("%s\n", f.FilePath))

		switch {
		case err != nil:
			execErr = m.cat.UnshareFile(&f.Sha1Hash)
			actionTaken.WriteString("  - File is missing. Removed from index\n")
			actionTaken.WriteString(fmt.Sprintf("  - %v\n", err))
		case !f.Progress.Full():
			actionTaken.WriteString("  - Download in progress. Ignored\n")
		case *currentHash != f.Sha1Hash:
			execErr = m.cat.UnshareFile(&f.Sha1Hash)
			actionTaken.WriteString("  - File has changed since indexing. Removed from Index\n")
			actionTaken.WriteString(fmt.Sprintf("    Indexed Hash: %s\n", f.Sha1Hash.String()))
			actionTaken.WriteString(fmt.Sprintf("    Current Hash: %s\n", currentHash.String()))
		default:
			actionTaken.WriteString("  - Download complete & Hashes match. Ignored\n")
		}

		if execErr != nil {
			return execErr
		}

		resp.Items = append(resp.Items, CleanResponseItem{
			FilePath:        f.FilePath,
			IndexedSha1Hash: f.Sha1Hash.Data,
			CurrentSha1Hash: currentHash.Data,
			ActionTaken:     actionTaken.String(),
		})
	}

	return nil
}
