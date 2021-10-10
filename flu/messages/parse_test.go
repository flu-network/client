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
	check(err)

	if !reflect.DeepEqual(result, msg) {
		t.Fatalf("msg does not match result. \nmsg:%v \nres:%v \n", msg, result)
	}
}

func TestDiscoverHostResponse(t *testing.T) {
	msg := &DiscoverHostResponse{
		Address:   [4]byte{192, 168, 86, 34},
		Port:      61690,
		RequestID: 45678,
	}

	serialized := msg.Serialize()
	result, err := Parse(serialized)
	check(err)

	if !reflect.DeepEqual(result, msg) {
		t.Fatalf("msg does not match result. \nmsg:%v \nres:%v \n", msg, result)
	}
}
