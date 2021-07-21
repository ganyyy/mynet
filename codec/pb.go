package codec

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/ganyyy/mynet"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	ErrMessageType = errors.New("error input message type. must proto.Message")
	ErrNotRegister = errors.New("not register proto.Message")
	ErrDupliateReg = errors.New("duplicate register message id")
	ErrReceiveID   = errors.New("receive msg id error")
	ErrPackageHead = errors.New("package head error")
	ErrMessageLen  = errors.New("package message len error")
)

//ProtoBufProtocol proto 对应的编码
type ProtoBufProtocol struct {
	idToProto map[uint16]protoreflect.MessageType // ID到类型的映射
	protoToId map[protoreflect.MessageType]uint16 // 类型到ID的映射
}

func PBProtocol() *ProtoBufProtocol {
	return &ProtoBufProtocol{
		idToProto: map[uint16]protoreflect.MessageType{},
		protoToId: map[protoreflect.MessageType]uint16{},
	}
}

func (p *ProtoBufProtocol) Register(id uint16, t proto.Message) error {
	if _, ok := p.idToProto[id]; ok {
		return ErrDupliateReg
	}
	var rt = t.ProtoReflect().Type()
	if _, ok := p.protoToId[rt]; ok {
		return ErrDupliateReg
	}
	p.idToProto[id] = rt
	p.protoToId[rt] = id
	return nil
}

func (p *ProtoBufProtocol) NewCodec(rw io.ReadWriter) (mynet.Codec, error) {
	return &pbCodec{
		p:         p,
		marshal:   &proto.MarshalOptions{},
		unmarshal: &proto.UnmarshalOptions{},
		rw:        rw,
	}, nil
}

type pbCodec struct {
	p         *ProtoBufProtocol
	marshal   *proto.MarshalOptions
	unmarshal *proto.UnmarshalOptions
	rw        io.ReadWriter
}

func (p *pbCodec) Receive() (interface{}, error) {
	var head [4]byte
	var n, err = io.ReadFull(p.rw, head[:])
	if n != 4 {
		return nil, ErrPackageHead
	}
	if err != nil {
		return nil, err
	}
	var size uint16
	var id uint16
	var ok bool
	size = binary.BigEndian.Uint16(head[:2])
	id = binary.BigEndian.Uint16(head[2:])

	var data = make([]byte, size)
	n, err = io.ReadFull(p.rw, data)
	if n != int(size) {
		return nil, ErrMessageLen
	}
	if err != nil {
		return nil, err
	}
	var pt protoreflect.MessageType
	pt, ok = p.p.idToProto[id]
	if !ok {
		return nil, ErrNotRegister
	}
	var pb = pt.New().Interface()

	err = p.unmarshal.Unmarshal(data, pb)
	return pb, err
}

func (p *pbCodec) Send(pb interface{}) error {
	var ok bool
	var pbMsg proto.Message
	if pbMsg, ok = pb.(proto.Message); !ok {
		return ErrMessageType
	}
	var id uint16
	if id, ok = p.p.protoToId[pbMsg.ProtoReflect().Type()]; !ok {
		return ErrNotRegister
	}
	var data, err = p.marshal.Marshal(pb.(proto.Message))
	if err != nil {
		return err
	}
	var head [4]byte
	binary.BigEndian.PutUint16(head[:2], uint16(len(data)))
	binary.BigEndian.PutUint16(head[2:], id)
	p.rw.Write(head[:])
	_, err = p.rw.Write(data)
	return err
}

func (p *pbCodec) Close() error {
	if closer, ok := p.rw.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
