package messages

import (
	"reflect"
	"testing"

	"github.com/flu-network/client/common"
)

func TestDiscoverHostRequest(t *testing.T) {
	h := common.Sha1Hash{}
	h.FromString("F10E2821BBBEA527EA02200352313BC059445190")
	msg := &DiscoverHostRequest{
		RequestID: 123,
		Sha1Hash:  h,
		Chunks:    []uint16{4, 5, 60123},
	}

	serialized := msg.Serialize()
	result, err := Parse(serialized)
	check(err, t)

	if !reflect.DeepEqual(result, msg) {
		t.Fatalf("msg does not match result. \nmsg:%v \nres:%v \n", msg, result)
	}
}

func TestDiscoverHostResponse(t *testing.T) {
	msg := &DiscoverHostResponse{
		Address:   [4]byte{192, 168, 86, 34},
		Port:      61690,
		RequestID: 45678,
		Chunks:    []uint16{},
	}

	serialized := msg.Serialize()
	result, err := Parse(serialized)
	check(err, t)

	if !reflect.DeepEqual(result, msg) {
		t.Fatalf("msg does not match result. \nmsg:%v \nres:%v \n", msg, result)
	}
}

func TestOpenLineRequest(t *testing.T) {
	h := common.Sha1Hash{}
	h.FromString("F10E2821BBBEA527EA02200352313BC059445190")
	msg := &OpenConnectionRequest{
		Sha1Hash:  &h,
		Chunk:     654,
		WindowCap: 213,
	}

	serialized := msg.Serialize()
	result, err := Parse(serialized)
	check(err, t)

	if !reflect.DeepEqual(result, msg) {
		t.Fatalf("msg does not match result. \nmsg:%v \nres:%v \n", msg, result)
	}
}

func TestListFilesRequest(t *testing.T) {
	h := common.Sha1Hash{}
	h.FromString("F10E2821BBBEA527EA02200352313BC059445190")
	msg := &ListFilesRequest{
		RequestID: 123,
		Sha1Hash:  &h,
	}

	serialized := msg.Serialize()
	result, err := Parse(serialized)
	check(err, t)

	if !reflect.DeepEqual(result, msg) {
		t.Fatalf("msg does not match result. \nmsg:%v \nres:%v \n", msg, result)
	}
}

func TestListFilesResponse(t *testing.T) {
	type testCase struct {
		desc  string
		input ListFilesResponse
	}

	testCases := []*testCase{
		{
			desc: "Empty",
			input: ListFilesResponse{
				RequestID: 123,
				Files:     []ListFilesEntry{},
			},
		},
		{
			desc: "Empty filename",
			input: ListFilesResponse{
				RequestID: 45678,
				Files: []ListFilesEntry{
					{
						SizeInBytes:      123,
						ChunkCount:       456,
						ChunkSizeInBytes: 7890,
						ChunksDownloaded: 7890,
						Sha1Hash:         (&common.Sha1Hash{}).FromString("F10E2821BBBEA527EA02200352313BC059445190"),
						FileName:         "",
					},
				},
			},
		},
		{
			desc: "full response",
			input: ListFilesResponse{
				RequestID: 45678,
				Files: []ListFilesEntry{
					{
						SizeInBytes:      123,
						ChunkCount:       456,
						ChunkSizeInBytes: 7890,
						ChunksDownloaded: 7890,
						Sha1Hash:         (&common.Sha1Hash{}).FromString("F10E2821BBBEA527EA02200352313BC059445190"),
						FileName:         "バットマン.rar",
					},
					{
						SizeInBytes:      1234,
						ChunkCount:       4567,
						ChunkSizeInBytes: 7890,
						ChunksDownloaded: 7800,
						Sha1Hash:         (&common.Sha1Hash{}).FromString("F10E2821BBBEA527EA02200352313BC059445190"),
						FileName:         "Batman_Begins (2017).mkv",
					},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			serialized := c.input.Serialize()
			result, err := Parse(serialized)
			check(err, t)
			if !reflect.DeepEqual(result, &c.input) {
				t.Fatalf("msg does not match result. \nmsg:%v \nres:%v \n", c.input, result)
			}
		})
	}
}

func check(e error, t *testing.T) {
	if e != nil {
		t.Fatal(e)
	}
}
