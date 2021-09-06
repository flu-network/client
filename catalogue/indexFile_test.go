package catalogue

import (
	"crypto/sha1"
	"reflect"
	"testing"
)

func TestMarshalling(t *testing.T) {
	// create a rich test subject
	subject := IndexFile{
		pid:         10293,
		lastTouched: 1630892423,
		index:       map[[20]byte]IndexRecord{},
	}
	subject.index[*sha1HashString("cat")] = IndexRecord{
		FilePath:     "path/to/file1.dat",
		SizeInBytes:  123456,
		Sha1Hash:     *sha1HashString("cat"),
		ProgressFile: nil,
	}
	subject.index[*sha1HashString("bat")] = IndexRecord{
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
	result := IndexFile{}
	result.UnmarshalJSON(data)

	// check they're exactly equal
	if !reflect.DeepEqual(subject, result) {
		t.Fatalf("Expected subject:\n%v\n to deep equal result:\n%v\n", subject, result)
	}
}

func sha1HashString(str string) *[20]byte {
	sha1A := sha1.New()
	sha1A.Write([]byte(str))
	resultSlice := sha1A.Sum(nil)
	return (*[20]byte)(resultSlice)
}
