package catalogue

import "github.com/flu-network/client/catalogue/bitset"

// ProgressFile is an in-memory representation of a file on disk containing a bitset, which shows
// which 'chunks' of a file has been downloaded. If the entire file has been downloaded, all bits
// in the set are 'on'.
type ProgressFile struct {
	Progress bitset.Bitset
}
