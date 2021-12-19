package catalogue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flu-network/client/common"
)

const indexFileName = "index.json"

// indexFile is the in-memory representation of the index. The index maps the sha1 hash of a file
// to the IndexRecord associated with that file. All methods assume the caller has acquired
// a mutex granting exclusive access.
type indexFile struct {
	// returns the pid of the process that created the file. The index file can be 'claimed' by the
	// running process if the pid matches OR the file has not been touched for 30 seconds. The
	// owner is expected update the file's lastTouched every few seconds (<< 30)
	pid         int
	lastTouched int64 // should be updated regularly by the owner
	index       map[common.Sha1Hash]*indexRecord
	dataDir     string
}

// Init attempts to safely claim ownnership of the index file if it already exists. If it
// does not exist, an index file is created.
func (ind *indexFile) Init(dataDir string) error {
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
			ind.index = map[common.Sha1Hash]*indexRecord{}
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

func (ind *indexFile) save() error {
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
func (ind *indexFile) AddIndexRecord(record *indexRecord) error {
	if extantRecord, exists := ind.index[record.Sha1Hash]; exists {
		return fmt.Errorf("identical file already shared: %s", extantRecord.FilePath)
	}
	ind.index[record.Sha1Hash] = record
	return ind.save()
}

// RemoveIndexRecord removes an indexRecord from the underlying file, and reloads the in-memory
// representation of the data so that change is reflected.
func (ind *indexFile) RemoveIndexRecord(record *indexRecord) error {
	delete(ind.index, record.Sha1Hash)
	return ind.save()
}

// MarshalJSON conforms to the Marshaler interface
func (ind *indexFile) MarshalJSON() ([]byte, error) {
	intermediary := indexFileJSON{
		Pid:         ind.pid,
		LastTouched: ind.lastTouched,
		Index:       map[string]indexRecordJSON{},
		DataDir:     ind.dataDir,
	}

	for hash, indexRecord := range ind.index {
		intermediary.Index[hash.String()] = *indexRecord.toJSON()
	}

	return json.Marshal(intermediary)
}

// UnmarshalJSON conforms to the Marshaler interface. Does not fill out pointers to the underlying
// ProgressFiles within individual indexRecords.
func (ind *indexFile) UnmarshalJSON(data []byte) error {
	intermediary := indexFileJSON{}
	err := json.Unmarshal(data, &intermediary)
	if err != nil {
		return err
	}

	ind.pid = intermediary.Pid
	ind.lastTouched = intermediary.LastTouched
	ind.index = make(map[common.Sha1Hash]*indexRecord)
	ind.dataDir = intermediary.DataDir

	for str, indexRecord := range intermediary.Index {
		hash := common.Sha1Hash{}
		err := hash.FromStringSafe(str)
		if err != nil {
			return err
		}

		decoded, err := indexRecord.fromJSON()
		if err != nil {
			return err
		}
		ind.index[hash] = decoded
	}

	return nil
}

// Sprint returns a pretty-printed string-representation of the in-memory indexFile
func (ind *indexFile) Sprint() string {
	var result strings.Builder
	// pid         int
	// lastTouched int64 // should be updated regularly by the owner
	// index       map[common.Sha1Hash]*indexRecord
	// dataDir     string

	result.WriteString(fmt.Sprintf("pid: %d\n", ind.pid))
	result.WriteString(fmt.Sprintf("lastTouched: %d\n", ind.lastTouched))
	result.WriteString(fmt.Sprintf("dataDir: %s\n", ind.dataDir))

	for k, rec := range ind.index {
		result.WriteString(fmt.Sprintf("  %v\n", k))
		result.WriteString(fmt.Sprintf("    %s: %s\n", "FilePath", rec.FilePath))
		result.WriteString(fmt.Sprintf("    %s: %s\n", "Sha1Hash", rec.Sha1Hash.String()))
	}

	return result.String()
}

// indexFileJSON is a private intermediary representation of an indexFile for JSON encoding
type indexFileJSON struct {
	Pid         int
	LastTouched int64
	Index       map[string]indexRecordJSON
	DataDir     string
}
