package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"

	"github.com/ganyyy/mynet"
)

var ErrTooLargePacket = errors.New("too large packet")

type FixLenProtocol struct {
	base        mynet.Protocol    // 编解码器
	n           int               // 编/解码的位数
	maxRecv     uint              // 最大接收包长度
	maxSend     uint              // 最大发送包长度
	headDecoder func([]byte) int  // 消息头的解码函数
	headEncoder func([]byte, int) // 消息头的编码函数
}

// 构建一个
func FixLen(base mynet.Protocol, n int, byteOrder binary.ByteOrder, maxRecv, maxSend uint) *FixLenProtocol {
	var getMin = func(a, b uint) uint {
		if a < b {
			return a
		}
		return b
	}

	var proto = &FixLenProtocol{
		n:    n,
		base: base,
	}
	switch n {
	case 1:
		maxRecv = getMin(maxRecv, math.MaxUint8)
		maxSend = getMin(maxSend, math.MaxUint8)
		proto.headDecoder = func(b []byte) int {
			return int(b[0])
		}
		proto.headEncoder = func(b []byte, i int) {
			b[0] = byte(i)
		}
	case 2:
		maxRecv = getMin(maxRecv, math.MaxUint16)
		maxSend = getMin(maxSend, math.MaxUint16)
		proto.headDecoder = func(b []byte) int {
			return int(byteOrder.Uint16(b))
		}
		proto.headEncoder = func(b []byte, i int) {
			byteOrder.PutUint16(b, uint16(i))
		}
	case 4:
		maxRecv = getMin(maxRecv, math.MaxUint32)
		maxSend = getMin(maxSend, math.MaxUint32)
		proto.headDecoder = func(b []byte) int {
			return int(byteOrder.Uint32(b))
		}
		proto.headEncoder = func(b []byte, i int) {
			byteOrder.PutUint32(b, uint32(i))
		}
	case 8:
		proto.headDecoder = func(b []byte) int {
			return int(byteOrder.Uint32(b))
		}
		proto.headEncoder = func(b []byte, i int) {
			byteOrder.PutUint32(b, uint32(i))
		}
	}

	proto.maxRecv = maxRecv
	proto.maxSend = maxSend

	return proto
}

func (f *FixLenProtocol) NewCodec(rw io.ReadWriter) (cc mynet.Codec, err error) {
	var codec = &fixLenCodec{
		rw:             rw,
		FixLenProtocol: f,
	}
	codec.headBuf = codec.head[:f.n]
	codec.base, err = f.base.NewCodec(&codec.fixLenReadWriter)
	return codec, err
}

type fixLenReadWriter struct {
	recvBuf bytes.Reader // 外来的数据写入到recvBuf中, 通过codec.Receive进行解码
	sendBuf bytes.Buffer // 内部的数据发送到sendBuf中, 通过codec.Send进行编码
}

func (rw *fixLenReadWriter) Read(p []byte) (int, error) {
	return rw.recvBuf.Read(p)
}

func (rw *fixLenReadWriter) Write(p []byte) (int, error) {
	return rw.sendBuf.Write(p)
}

type fixLenCodec struct {
	*FixLenProtocol
	fixLenReadWriter

	base    mynet.Codec
	head    [8]byte // 头部最长八字节.
	headBuf []byte  // 读取头部用的缓冲区
	bodyBuf []byte  // 读取消息用的缓冲区
	rw      io.ReadWriter
}

//Receive 消息读取
func (f *fixLenCodec) Receive() (interface{}, error) {
	// 读取头部长度
	if _, err := io.ReadFull(f.rw, f.headBuf); err != nil {
		return nil, err
	}
	var size = f.headDecoder(f.headBuf)
	if size > int(f.maxRecv) {
		return nil, ErrTooLargePacket
	}
	// 分配足够的空间来接收数据
	if cap(f.bodyBuf) < size {
		f.bodyBuf = make([]byte, size, size+128)
	}
	var buff = f.bodyBuf[:size]
	if _, err := io.ReadFull(f.rw, buff); err != nil {
		return nil, err
	}
	f.recvBuf.Reset(buff)
	msg, err := f.base.Receive()
	return msg, err
}

//Send 消息发送
func (f *fixLenCodec) Send(msg interface{}) error {
	f.sendBuf.Reset()
	// 预写入包头空间
	f.sendBuf.Write(f.headBuf)
	err := f.base.Send(msg)
	if err != nil {
		return err
	}
	buff := f.sendBuf.Bytes()
	// 对包头空间进行编码
	f.headEncoder(buff, len(buff)-f.n)
	_, err = f.rw.Write(buff)
	return err
}

//Close 关闭解码器
func (f *fixLenCodec) Close() error {
	if closer, ok := f.rw.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
