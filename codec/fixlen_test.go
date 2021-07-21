package codec_test

import (
	"encoding/binary"
	"testing"

	"github.com/ganyyy/mynet/codec"
)

func TestFixLen(t *testing.T) {
	base := JsonTestProtocol()
	protocol := codec.FixLen(base, 2, binary.BigEndian, 1024, 1024)
	JsonTest(t, protocol)
}

func TestFixLen2(t *testing.T) {
	base := PBTestProtocol()
	protocol := codec.FixLen(base, 2, binary.BigEndian, 1024, 1024)
	PBTest(t, protocol)
}

func TestFixLen3(t *testing.T) {
	base := JsonTestProtocol()
	proto := codec.FixLen(base, 2, binary.BigEndian, 1024, 1024)
	JsonTest2(t, proto)
}
