package codec_test

import (
	"mynet/proto/demo"
	"sync"
	"testing"

	"github.com/ganyyy/mynet"
	"github.com/ganyyy/mynet/codec"
)

func PBTestProtocol() *codec.ProtoBufProtocol {
	var pb = codec.PBProtocol()
	pb.Register(1, &demo.Req{})
	pb.Register(2, &demo.Rsp{})
	return pb
}

func PBTest(t *testing.T, protocol mynet.Protocol) {
	var stream safeBuffer

	codec, _ := protocol.NewCodec(&stream)

	var sendMsg = &demo.Req{
		Str: "hello world",
	}

	var wait sync.WaitGroup

	const (
		SendNum = 10
		RecvNum = 10
	)

	wait.Add(2)
	go func() {
		defer wait.Done()
		for i := 0; i < SendNum; i++ {
			codec.Send(sendMsg)
		}
	}()
	go func() {
		defer wait.Done()
		for i := 0; i < RecvNum; i++ {
			info, err := codec.Receive()
			t.Logf("[%v]:%v", info, err)
		}
	}()

	wait.Wait()
}

func TestPBProto(t *testing.T) {
	PBTest(t, PBTestProtocol())
}
