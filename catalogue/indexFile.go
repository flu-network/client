package catalogue

import (
	"encoding/hex"
	"encoding/json"
)

// IndexFile is the in-memory representation of the index. The index maps the sha1 hash of a file
// to the IndexRecord associated with that file.
type IndexFile struct {
	// returns the pid of the process that created the file. The index file can be 'claimed' by the
	// running process if the pid matches OR the file has not been touched for 30 seconds. The
	// owner is expected update the file's lastTouched every few seconds (<< 30)
	pid         int
	lastTouched int64 // should be updated regularly by the owner
	index       map[[20]byte]IndexRecord
}

// indexFileJSON is a private intermediary representation of an IndexFile for JSON encoding
type indexFileJSON struct {
	Pid         int
	LastTouched int64
	Index       map[string]IndexRecord
}

// MarshalJSON conforms to the Marshaler interface
func (ind *IndexFile) MarshalJSON() ([]byte, error) {
	intermediary := indexFileJSON{
		Pid:         ind.pid,
		LastTouched: ind.lastTouched,
		Index:       map[string]IndexRecord{},
	}

	for hash, indexRecord := range ind.index {
		strHash := hex.EncodeToString(hash[:])
		intermediary.Index[strHash] = indexRecord
	}

	return json.Marshal(intermediary)
}

// UnmarshalJSON conforms to the Marshaler interface
func (ind *IndexFile) UnmarshalJSON(data []byte) error {
	intermediary := indexFileJSON{}
	err := json.Unmarshal(data, &intermediary)
	if err != nil {
		return err
	}

	ind.pid = intermediary.Pid
	ind.lastTouched = intermediary.LastTouched
	ind.index = make(map[[20]byte]IndexRecord)

	for str, indexRecord := range intermediary.Index {
		strHash, err := hex.DecodeString(str)
		if err != nil {
			return err
		}
		strHashArray := [20]byte{}
		copy(strHashArray[:], strHash)
		ind.index[strHashArray] = indexRecord
	}

	return nil
}
