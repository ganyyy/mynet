package codec

import (
	"encoding/json"
	"io"
	"reflect"

	"github.com/ganyyy/mynet"
)

type JsonProtocol struct {
	strToType map[string]reflect.Type
	typeToStr map[reflect.Type]string
}

func Json() *JsonProtocol {
	return &JsonProtocol{
		strToType: make(map[string]reflect.Type),
		typeToStr: make(map[reflect.Type]string),
	}
}

func (j *JsonProtocol) Register(t interface{}) {
	var rt = reflect.TypeOf(t)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	var name = rt.PkgPath() + "/" + rt.Name()
	j.typeToStr[rt] = name
	j.strToType[name] = rt
}

func (j *JsonProtocol) RegisterNmae(name string, t interface{}) {
	var rt = reflect.TypeOf(t)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	j.typeToStr[rt] = name
	j.strToType[name] = rt
}

func (j *JsonProtocol) NewCodec(rw io.ReadWriter) (mynet.Codec, error) {
	var codec = &jsonCodec{
		p:      j,
		encode: json.NewEncoder(rw),
		decode: json.NewDecoder(rw),
	}

	codec.closer, _ = rw.(io.Closer)
	return codec, nil
}

type jsonCodec struct {
	p      *JsonProtocol
	closer io.Closer
	encode *json.Encoder
	decode *json.Decoder
}

type jsonIn struct {
	Head string
	Body *json.RawMessage
}

type jsonOut struct {
	Head string
	Body interface{}
}

func (j *jsonCodec) Receive() (interface{}, error) {
	var in jsonIn
	err := j.decode.Decode(&in)
	if err != nil {
		return nil, err
	}

	var body interface{}
	if in.Head != "" {
		if t, exist := j.p.strToType[in.Head]; exist {
			body = reflect.New(t).Interface()
		}
	}
	err = json.Unmarshal(*in.Body, body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (j *jsonCodec) Send(msg interface{}) error {
	var out jsonOut
	var t = reflect.TypeOf(msg)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if name, exist := j.p.typeToStr[t]; exist {
		out.Head = name
	}
	out.Body = msg
	return j.encode.Encode(out)
}

func (j *jsonCodec) Close() error {
	if j.closer != nil {
		return j.closer.Close()
	}
	return nil
}
