package catalogue

import (
	"crypto/sha1"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/flu-network/client/common"
)

func TestMarshalling(t *testing.T) {
	// create a rich test subject
	subject := indexFile{
		pid:         10293,
		lastTouched: 1630892423,
		index:       map[common.Sha1Hash]*indexRecord{},
	}
	subject.index[*sha1HashString("cat")] = &indexRecord{
		FilePath:     "path/to/file1.dat",
		SizeInBytes:  123456,
		Sha1Hash:     *sha1HashString("cat"),
		ProgressFile: nil,
	}
	subject.index[*sha1HashString("bat")] = &indexRecord{
		FilePath:     "path/to/file2.mkv",
		SizeInBytes:  13243546,
		Sha1Hash:     *sha1HashString("bat"),
		ProgressFile: nil,
	}

	// serialize it
	data, err := subject.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	// deserialize it
	result := indexFile{}
	result.UnmarshalJSON(data)

	// check they're exactly equal
	if !reflect.DeepEqual(subject, result) {
		t.Fatalf("Expected subject:\n%v\n to deep equal result:\n%v\n", subject, result)
	}
}

func TestInit(t *testing.T) {
	dataDir := filepath.Join(string(os.PathSeparator), "tmp", "flu-client", "catalogue")
	parentDir := filepath.Join(string(os.PathSeparator), "tmp", "flu-client")
	var cleanup = func() {
		err := os.RemoveAll(parentDir)
		if err != nil {
			panic(err)
		}
	}

	t.Run("Works when file does not already exist", func(t *testing.T) {
		cleanup()
		defer cleanup()
		indexFile := &indexFile{}
		err := indexFile.Init(dataDir)

		if err != nil {
			t.Fatal(err)
		}

		now := time.Now().Unix()
		if now-indexFile.lastTouched > 1000 {
			t.Fatalf(
				"Expected lastTouched to roughly match but %d is way over %d\n",
				now, indexFile.lastTouched,
			)
		}

		pid := os.Getpid()
		if pid != indexFile.pid {
			t.Fatalf("Expected pid to match but %d != %d\n", pid, indexFile.pid)
		}

		if len(indexFile.index) != 0 {
			t.Fatalf("Expected fresh indexFile to have empty index but it wasn't\n")
		}
	})

	// TODO: cover the unhappy cases
}

func sha1HashString(str string) *common.Sha1Hash {
	sha1A := sha1.New()
	sha1A.Write([]byte(str))
	result := common.Sha1Hash{}
	return result.FromSlice(sha1A.Sum(nil))
}
