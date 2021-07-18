package codec_test

import (
	"bytes"
	"testing"

	"github.com/ganyyy/mynet"
	"github.com/ganyyy/mynet/codec"
)

type MyMessage1 struct {
	Field1 string
	Field2 int
}

type MyMessage2 struct {
	Field1 int
	Field2 string
}

func JsonTestProtocol() *codec.JsonProtocol {
	var protocol = codec.Json()
	protocol.Register(MyMessage1{})
	protocol.RegisterNmae("msg2", MyMessage2{})
	return protocol
}

func JsonTest(t *testing.T, protocol mynet.Protocol) {

	var stream bytes.Buffer

	codec, _ := protocol.NewCodec(&stream)

	var sendMsg1 = MyMessage1{
		Field1: "123",
		Field2: 456,
	}

	err := codec.Send(&sendMsg1)
	if err != nil {
		t.Fatalf("sendMsg1 error:%v", err)
	}

	t.Logf("cur stream:%v", stream.String())

	recvMsg1, err := codec.Receive()
	if err != nil {
		t.Fatalf("recvMsg1 error:%v", err)
	}

	if _, ok := recvMsg1.(*MyMessage1); !ok {
		t.Fatalf("receive message type not")
	} else {
		t.Logf("receive message1:%v", recvMsg1.(*MyMessage1))
	}

	var sendMsg2 = MyMessage2{
		Field1: 1000,
		Field2: "123",
	}

	err = codec.Send(&sendMsg2)
	if err != nil {
		t.Fatal(err)
	}

	recvMsg2, err := codec.Receive()
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := recvMsg2.(*MyMessage2); !ok {
		t.Fatalf("message type not match: %#v", recvMsg2)
	}

	if sendMsg2 != *(recvMsg2.(*MyMessage2)) {
		t.Fatalf("message not match %v, %v", sendMsg1, recvMsg1)
	}
}

func TestJsonProtocol(t *testing.T) {
	JsonTest(t, JsonTestProtocol())
}
