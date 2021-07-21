package codec_test

import (
	"bytes"
	"sync"
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

type safeBuffer struct {
	bytes.Buffer
	lock sync.Mutex
}

func (s *safeBuffer) Read(buf []byte) (n int, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.Buffer.Read(buf)
}

func (s *safeBuffer) Write(buf []byte) (n int, err error) {
	s.lock.Lock()
	n, err = s.Buffer.Write(buf)
	s.lock.Unlock()
	return
}

func JsonTest2(t *testing.T, protocol mynet.Protocol) {
	// bytes.Buffer 不适合并发读写
	var stream safeBuffer

	codec, _ := protocol.NewCodec(&stream)

	var wait sync.WaitGroup
	const (
		SendNum = 50
		RecvNum = 50
	)

	wait.Add(SendNum + RecvNum)
	var sendMsg1 = MyMessage1{
		Field1: "123",
		Field2: 456,
	}
	for i := 0; i < SendNum; i++ {
		go func() {
			defer wait.Done()
			codec.Send(sendMsg1)
		}()
	}

	for i := 0; i < RecvNum; i++ {
		go func() {
			defer wait.Done()
			info, err := codec.Receive()
			t.Logf("[%v]:%v", info, err)
		}()
	}

	wait.Wait()
}

func TestJsonProtocol(t *testing.T) {
	JsonTest(t, JsonTestProtocol())
}
