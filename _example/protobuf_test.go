package eudore_test

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/eudore/eudore"
)

func TestProtoBuf(*testing.T) {
	type Data struct {
		Bool      bool              `protobuf:"varint,1,opt,name=Bool,proto3" json:"Bool,omitempty"`
		Int64     int64             `protobuf:"varint,2,opt,name=Int64,proto3" json:"Int64,omitempty"`
		Uint64    uint64            `protobuf:"varint,3,opt,name=Uint64,proto3" json:"Uint64,omitempty"`
		Double    float64           `protobuf:"fixed64,4,opt,name=Double,proto3" json:"Double,omitempty"`
		Float     float32           `protobuf:"fixed32,5,opt,name=Float,proto3" json:"Float,omitempty"`
		String    string            `protobuf:"bytes,6,opt,name=String,proto3" json:"String,omitempty"`
		Bytes     []byte            `protobuf:"bytes,7,opt,name=Bytes,proto3" json:"Bytes,omitempty"`
		Time      time.Time         `protobuf:"bytes,16,opt,name=Time,proto3" json:"Time,omitempty"`
		MapString map[int64]string  `protobuf:"bytes,9,rep,name=MapString,proto3" json:"MapString,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
		MapBool   map[int64]bool    `protobuf:"bytes,10,rep,name=MapBool,proto3" json:"MapBool,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
		MapFloat  map[int64]float32 `protobuf:"bytes,11,rep,name=MapFloat,proto3" json:"MapFloat,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"fixed32,2,opt,name=value,proto3"`
		Func      func()            `protobuf:"varint,12,opt,name=Func,proto3" json:"-"`
		None      int               `protobuf:"-"`
	}

	type Slice struct {
		Bool   []bool      `protobuf:"varint,1,rep,packed,name=Bool,proto3" json:"Bool,omitempty"`
		Int64  []int64     `protobuf:"varint,2,rep,packed,name=Int64,proto3" json:"Int64,omitempty"`
		Uint64 []uint64    `protobuf:"varint,3,rep,packed,name=Uint64,proto3" json:"Uint64,omitempty"`
		Double []float64   `protobuf:"fixed64,4,rep,packed,name=Double,proto3" json:"Double,omitempty"`
		Float  []float32   `protobuf:"fixed32,5,rep,packed,name=Float,proto3" json:"Float,omitempty"`
		String []string    `protobuf:"bytes,6,rep,name=String,proto3" json:"String,omitempty"`
		Bytes  [][]byte    `protobuf:"bytes,7,rep,name=Bytes,proto3" json:"Bytes,omitempty"`
		Time   []time.Time `protobuf:"bytes,16,rep,name=Time,proto3" json:"Time,omitempty"`
		Datas  []*Data     `protobuf:"bytes,10,rep,name=Datas,proto3" json:"Datas,omitempty"`
		Data   *Data       `protobuf:"bytes,12,opt,name=Data,proto3" json:"Data,omitempty"`
	}

	times := []time.Time{
		time.Unix(1257894000, 692832033),
		time.Unix(1666122841, 0),
		time.Unix(1666143873, 123974933),
		time.Unix(1666158958, 569372220),
	}

	data := &Data{
		Bool:      true,
		Int64:     -0x6000000000000000,
		Uint64:    0xffffffffffffffff,
		Double:    83.224,
		Float:     82.833,
		String:    "eudore",
		Bytes:     []byte("protobuf"),
		MapString: map[int64]string{1: "eudore"},
		MapBool:   map[int64]bool{4: true},
		MapFloat:  map[int64]float32{4: 888.3333},
		Time:      times[0],
	}

	slice := &Slice{
		Bool:   []bool{true, false, false, true},
		Int64:  []int64{1, 2, 3, 4, -0x6000000000000000},
		Uint64: []uint64{1, 2, 3, 4, 0xffffffffffffffff},
		Double: []float64{0xffffffffffffffff, 9999999999.9999999999},
		Float:  []float32{0xffffffffffffffff, 9999999999.9999999999, -11222.123456789},
		String: []string{"strings1", "length", "is", "4"},
		Bytes:  [][]byte{[]byte("eudore"), []byte("struct"), []byte("protobuf"), []byte("encoding")},
		Time:   times,
		Datas:  []*Data{data, data},
		Data:   data,
	}

	fmt.Println("-------------encode-------------")
	hash1, body1 := encodeEudore(data)
	hash2, body2 := encodeEudore(slice)
	// encodeEudore([]*Data)

	fmt.Println("-------------decode-------------")
	data3 := &Data{}
	slice3 := &Slice{}
	decodeEudore(data3, body1)
	decodeEudore(slice3, body2)
	hash3, _ := encodeEudore(data3)
	hash4, _ := encodeEudore(slice3)
	fmt.Println(hash1 == "7bc9efbb008a6de3aea47254626ffcbf", hash2 == "6ad88cfe278a0b0f46b9376628cf9d8e")
	fmt.Println(hash3 == "7bc9efbb008a6de3aea47254626ffcbf", hash4 == "6ad88cfe278a0b0f46b9376628cf9d8e")
}

func encodeEudore(i interface{}) (string, []byte) {
	buf := &bytes.Buffer{}
	encoder := eudore.NewProtobufEncoder(buf)
	size := encoder.SizeMessage(i)
	encoder.Encode(i)

	h := md5.New()
	h.Write(buf.Bytes())
	hash := hex.EncodeToString(h.Sum(nil))
	fmt.Println(hash, size, buf.Len(), buf.Bytes())
	return hash, buf.Bytes()
}

func decodeEudore(i interface{}, body []byte) error {
	return eudore.NewProtobufDecoder(bytes.NewReader(body)).Decode(i)
}

func TestProtoBufEncoder(t *testing.T) {
	type Data struct {
		Name    string
		MapFunc map[string]func(*testing.T)
	}
	data := &Data{
		Name: "eudore funcs",
		MapFunc: map[string]func(*testing.T){
			"TestProtoBufEncoder": TestProtoBufEncoder,
			"TestProtoBufDecoder": TestProtoBufDecoder,
		},
	}
	buf := &bytes.Buffer{}
	encoder := eudore.NewProtobufEncoder(buf)
	size := encoder.SizeMessage(data)
	encoder.Encode(data)
	t.Log(size, buf.Len(), buf.Bytes())
}

func TestProtoBufDecoder(t *testing.T) {
	type Data struct {
		Bool       []bool           `protobuf:"varint,1,rep,packed,name=Bool,proto3" json:"Bool,omitempty"`
		String     string           `protobuf:"bytes,6,opt,name=String,proto3" json:"String,omitempty"`
		Time       time.Time        `protobuf:"bytes,16,opt,name=Time,proto3" json:"Time,omitempty"`
		MapString  map[int64]string `protobuf:"bytes,9,rep,name=MapString,proto3" json:"MapString,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
		MapBool    map[int64]bool   `protobuf:"bytes,10,rep,name=MapBool,proto3" json:"MapBool,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
		Interface  interface{}      `protobuf:"bytes,11,opt,name=Interface,proto3" json:"Interface,omitempty"`
		Interfaces []interface{}    `protobuf:"bytes,12,opt,name=Interfaces,proto3" json:"Interfaces,omitempty"`
		Funcs      []func()         `protobuf:"bytes,13,opt,name=Funcs,proto3" json:"Funcs,omitempty"`
		Stringers  []*fmt.Stringer  `protobuf:"bytes,14,opt,name=Stringers,proto3" json:"Stringers,omitempty"`
	}
	type Slice struct {
		Datas []*Data `protobuf:"bytes,10,rep,name=Datas,proto3" json:"Datas,omitempty"`
		Data  *Data   `protobuf:"bytes,12,opt,name=Data,proto3" json:"Data,omitempty"`
	}

	datas := [][]byte{
		{96}, // encode type struct invalid flag 0
		// message
		{98}, // decode message read len error: EOF
		{98, 6, 50, 6, 108, 101, 110, 103, 116, 104},  // decode message invalid len: 6, have data 4
		{98, 10, 50, 6, 108, 101, 110, 103, 116, 104}, // decode message has 2 bytes no read
		{98, 10, 50, 8, 108, 101, 110, 103, 116, 104}, // read length 8 has 6
		// slice
		{98, 8, 98, 6, 108, 101, 110, 103, 116, 104},
		{98, 8, 106, 6, 108, 101, 110, 103, 116, 104},
		// struct
		{98, 8, 7, 6, 108, 101, 110, 103, 116, 104}, // struct invalid flag 7
		{98, 15, 130, 1, 12, 8, 240, 224, 231, 215, 4, 16, 161, 142, 175, 202, 2},
		{98, 15, 130, 1, 12},     // decode struct time read flag error: EOF
		{98, 15, 130, 1, 12, 10}, // decode struct time invalid flag 10
		{98, 15, 130, 1, 12, 8},  // decode struct time read message error: EOF
		// struct discard
		{98, 11, 16, 128, 128, 128, 128, 128, 128, 128, 128, 160, 1},
		{98, 9, 17, 0, 0, 0, 0, 0, 0, 240, 67},
		{98, 8, 18, 6, 108, 101, 110, 103, 116, 104},
		{98, 5, 21, 126, 88, 47, 198},
		{98, 8, 16},    // decode struct discard read varint error: EOF
		{98, 8, 17},    // decode struct discard read float64 error: EOF
		{98, 8, 18},    // decode struct discard read len error: EOF
		{98, 8, 18, 8}, // decode struct discard read len invliad 8
		{98, 8, 18, 6}, // decode struct discard read message error: EOF
		{98, 8, 21},    // decode struct discard read float32 error: EOF
		{98, 8, 22},    // decode struct discard invalid flag 22
		// map
		{98, 12, 74, 10, 8, 1, 18, 6, 101, 117, 100, 111, 114, 101},
		{98, 12, 74, 10},    // decode size has 10 bytes no read
		{98, 12, 74, 10, 8}, // decode map read value error: EOF
		{98, 12, 74, 10, 24, 1, 18, 6, 101, 117, 100, 111, 114, 101}, // decode map invalid flag 24
	}
	for i := range datas {
		slice := &Slice{}
		t.Log(i, eudore.NewProtobufDecoder(bytes.NewReader(datas[i])).Decode(slice))
	}

	var str string
	var ster fmt.Stringer
	data1 := []byte{90, 4, 101, 117, 100, 111}
	t.Log(eudore.NewProtobufEncoder(io.Discard).Encode([]Data{}))
	t.Log(eudore.NewProtobufDecoder(bytes.NewReader(data1)).Decode([]Data{}))
	t.Log(eudore.NewProtobufDecoder(bytes.NewReader(data1)).Decode(&Data{}))                                                                      // decodeFlag nil interface
	t.Log(eudore.NewProtobufDecoder(bytes.NewReader(data1)).Decode(&Data{Interface: &str}))                                                       // decodeFlag
	t.Log(eudore.NewProtobufDecoder(bytes.NewReader([]byte{98, 6, 108, 101, 110, 103, 116, 104})).Decode(&Data{}))                                // decode slice interface 0 is nil
	t.Log(eudore.NewProtobufDecoder(bytes.NewReader([]byte{98, 6, 108, 101, 110, 103, 116, 104})).Decode(&Data{Interfaces: []interface{}{}}))     // decodeSlice
	t.Log(eudore.NewProtobufDecoder(bytes.NewReader([]byte{98, 6, 108, 101, 110, 103, 116, 104})).Decode(&Data{Interfaces: []interface{}{&str}})) // decodeSlice

	t.Log(eudore.NewProtobufDecoder(bytes.NewReader([]byte{114, 6, 108, 101, 110, 103, 116, 104})).Decode(&Data{Stringers: []*fmt.Stringer{&ster}})) // decodeSlice

	t.Log(eudore.NewProtobufDecoder(bytes.NewReader([]byte{82, 8, 50, 6, 108, 101, 110, 103, 116, 104})).Decode(&Slice{}))
	t.Log(eudore.NewProtobufDecoder(bytes.NewReader([]byte{82, 8, 50, 6, 108, 101, 110, 103, 116, 104})).Decode(&Slice{Datas: []*Data{{}}}))
}

func TestProtoBufHandlerData(t *testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/*", func(ctx eudore.Context) interface{} {
		type Data struct {
			Bool   bool    `protobuf:"varint,1,opt,name=Bool,proto3" json:"Bool,omitempty"`
			Int64  int64   `protobuf:"varint,2,opt,name=Int64,proto3" json:"Int64,omitempty"`
			Uint64 uint64  `protobuf:"varint,3,opt,name=Uint64,proto3" json:"Uint64,omitempty"`
			Double float64 `protobuf:"fixed64,4,opt,name=Double,proto3" json:"Double,omitempty"`
			Float  float32 `protobuf:"fixed32,5,opt,name=Float,proto3" json:"Float,omitempty"`
			String string  `protobuf:"bytes,6,opt,name=String,proto3" json:"String,omitempty"`
		}
		data := &Data{
			String: "eudore",
		}
		ctx.Bind(data)
		return data
	})

	app.NewRequest(nil, "PUT", "/",
		eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationProtobuf),
		eudore.NewClientHeader(eudore.HeaderContentType, eudore.MimeApplicationProtobuf),
		eudore.NewClientBody(nil),
		eudore.NewClientDumpHead(),
	)

	app.CancelFunc()
	app.Run()
}

/*
syntax = "proto3";
package main;

message Data {
bool Bool=1;
int64 Int64=2;
uint64 Uint64=3;
double Double=4;
float Float=5;
string String=6;
bytes Bytes=7;
Timestamp Time=16;
map<int64,string> MapString=9;
map<int64,bool> MapBool=10;
map<int64,float> MapFloat=11;
}

message Slice {
repeated bool Bool=1;
repeated int64 Int64=2;
repeated uint64 Uint64=3;
repeated double Double=4;
repeated float Float=5;
repeated string String=6;
repeated bytes Bytes=7;
repeated Timestamp Time=16;
repeated Data Datas=10;
Data Data=12;
}

message Timestamp {
int64 seconds = 1;
int32 nanos = 2;
}

7bc9efbb008a6de3aea47254626ffcbf
8 1
16 128 128 128 128 128 128 128 128 160 1
24 255 255 255 255 255 255 255 255 255 1
33 117 147 24 4 86 206 84 64
45 127 170 165 66
50 6 101 117 100 111 114 101
58 8 112 114 111 116 111 98 117 102
74 10 8 1 18 6 101 117 100 111 114 101
82 4 8 4 16 1
90 7 8 4 21 85 21 94 68
130 1 12 8 240 224 231 215 4 16 161 142 175 202 2

6ad88cfe278a0b0f46b9376628cf9d8e
10 4 1 0 0 1
18 14 1 2 3 4 128 128 128 128 128 128 128 128 160 1
26 14 1 2 3 4 255 255 255 255 255 255 255 255 255 1
34 16 0 0 0 0 0 0 240 67 0 0 0 32 95 160 2 66
42 12 0 0 128 95 249 2 21 80 126 88 47 198
50 8 115 116 114 105 110 103 115 49
50 6 108 101 110 103 116 104
50 2 105 115
50 1 52
58 6 101 117 100 111 114 101
58 6 115 116 114 117 99 116
58 8 112 114 111 116 111 98 117 102
58 8 101 110 99 111 100 105 110 103
82 98
	8 1
	16 128 128 128 128 128 128 128 128 160 1
	24 255 255 255 255 255 255 255 255 255 1
	33 117 147 24 4 86 206 84 64
	45 127 170 165 66
	50 6 101 117 100 111 114 101
	58 8 112 114 111 116 111 98 117 102
	74 10 8 1 18 6 101 117 100 111 114 101
	82 4 8 4 16 1
	90 7 8 4 21 85 21 94 68
	130 1 12 8 240 224 231 215 4 16 161 142 175 202 2
82 98 8 1 16 128 128 128 128 128 128 128 128 160 1 24 255 255 255 255 255 255 255 255 255 1 33 117 147 24 4 86 206 84 64 45 127 170 165 66 50 6 101 117 100 111 114 101 58 8 112 114 111 116 111 98 117 102 74 10 8 1 18 6 101 117 100 111 114 101 82 4 8 4 16 1 90 7 8 4 21 85 21 94 68 130 1 12 8 240 224 231 215 4 16 161 142 175 202 2
98 98 8 1 16 128 128 128 128 128 128 128 128 160 1 24 255 255 255 255 255 255 255 255 255 1 33 117 147 24 4 86 206 84 64 45 127 170 165 66 50 6 101 117 100 111 114 101 58 8 112 114 111 116 111 98 117 102 74 10 8 1 18 6 101 117 100 111 114 101 82 4 8 4 16 1 90 7 8 4 21 85 21 94 68 130 1 12 8 240 224 231 215 4 16 161 142 175 202 2
130 1 12 8 240 224 231 215 4 16 161 142 175 202 2
130 1 6 8 217 136 188 154 6
130 1 11 8 129 173 189 154 6 16 149 234 142 59
130 1 12 8 238 162 190 154 6 16 188 220 191 143 2]
*/
