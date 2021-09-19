package catalogue

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const indexFileName = "index.json"

// IndexFile is the in-memory representation of the index. The index maps the sha1 hash of a file
// to the IndexRecord associated with that file. All methods assume the caller has acquired
// a mutex granting exclusive access.
type IndexFile struct {
	// returns the pid of the process that created the file. The index file can be 'claimed' by the
	// running process if the pid matches OR the file has not been touched for 30 seconds. The
	// owner is expected update the file's lastTouched every few seconds (<< 30)
	pid         int
	lastTouched int64 // should be updated regularly by the owner
	index       map[[20]byte]IndexRecord
	dataDir     string
}

// Init attempts to safely claim ownnership of the index file if it already exists. If it
// does not exist, an index file is created.
func (ind *IndexFile) Init(dataDir string) error {
	// ensure directory exists
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return err
	}

	// check the container is accessible and is a directory
	fileInfo, err := os.Stat(dataDir)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("DataDir '%s' is not a directory", dataDir)
	}

	// ensure the file exists
	indexFilePath := filepath.Join(dataDir, indexFileName)
	_, err = os.Stat(indexFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// create it if it doesn't exist
			ind.pid = os.Getpid()
			ind.lastTouched = time.Now().Unix()
			ind.index = map[[20]byte]IndexRecord{}
			ind.dataDir = dataDir

			err := ind.save()
			if err != nil {
				return err
			}
		} else {
			// if there's some other error just bail out
			return err
		}
	}

	data, err := os.ReadFile(indexFilePath)
	if err != nil {
		return err
	}

	err = ind.UnmarshalJSON(data)
	// TODO: assert ownership
	return err
}

func (ind *IndexFile) save() error {
	data, err := json.Marshal(ind)
	if err != nil {
		return err
	}

	indexFilePath := filepath.Join(ind.dataDir, indexFileName)
	err = os.WriteFile(indexFilePath, data, 0664)
	if err != nil {
		return err
	}

	return nil
}

// AddIndexRecord adds an indexRecord to the underlying file, and reloads the in-memory
// representation of the data so that change is reflected. If an identical file has already been
// shared, this will safely return an error
func (ind *IndexFile) AddIndexRecord(record *IndexRecord) error {
	if extantRecord, exists := ind.index[record.Sha1Hash]; exists {
		return fmt.Errorf("Identical file already shared: %s", extantRecord.FilePath)
	}
	ind.index[record.Sha1Hash] = *record
	return ind.save()
}

// MarshalJSON conforms to the Marshaler interface
func (ind *IndexFile) MarshalJSON() ([]byte, error) {
	intermediary := indexFileJSON{
		Pid:         ind.pid,
		LastTouched: ind.lastTouched,
		Index:       map[string]indexRecordJSON{},
		DataDir:     ind.dataDir,
	}

	for hash, indexRecord := range ind.index {
		strHash := hex.EncodeToString(hash[:])
		intermediary.Index[strHash] = *indexRecord.toJSON()
	}

	return json.Marshal(intermediary)
}

// UnmarshalJSON conforms to the Marshaler interface. Does not fill out pointers to the underlying
// ProgressFiles within individual IndexRecords.
func (ind *IndexFile) UnmarshalJSON(data []byte) error {
	intermediary := indexFileJSON{}
	err := json.Unmarshal(data, &intermediary)
	if err != nil {
		return err
	}

	ind.pid = intermediary.Pid
	ind.lastTouched = intermediary.LastTouched
	ind.index = make(map[[20]byte]IndexRecord)
	ind.dataDir = intermediary.DataDir

	for str, indexRecord := range intermediary.Index {
		strHash, err := hex.DecodeString(str)
		if err != nil {
			return err
		}

		strHashArray := [20]byte{}
		copy(strHashArray[:], strHash)

		decoded, err := indexRecord.fromJSON()
		if err != nil {
			return err
		}
		ind.index[strHashArray] = *decoded
	}

	return nil
}

// indexFileJSON is a private intermediary representation of an IndexFile for JSON encoding
type indexFileJSON struct {
	Pid         int
	LastTouched int64
	Index       map[string]indexRecordJSON
	DataDir     string
}
