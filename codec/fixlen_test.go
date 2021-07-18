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
