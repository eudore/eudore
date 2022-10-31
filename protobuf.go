package eudore

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ProtobufEncoder 定义protobuf编码器
type ProtobufEncoder struct {
	io.Writer
	structFields map[reflect.Type][]protobufStructField
	varintBuffer [10]byte
}

// NewProtobufEncoder 函数创建一个protobuf编码器。
func NewProtobufEncoder(w io.Writer) *ProtobufEncoder {
	return &ProtobufEncoder{
		Writer:       w,
		structFields: make(map[reflect.Type][]protobufStructField),
	}
}

// Encode 方法执行probubuf编码，会从结构体和tag中获取对应的proto信息。
func (enc *ProtobufEncoder) Encode(i interface{}) error {
	iValue := reflect.Indirect(reflect.ValueOf(i))
	if iValue.Kind() != reflect.Struct {
		return fmt.Errorf(ErrFormatProtobufTypeMustSturct, reflect.TypeOf(i).String())
	}
	enc.encode(iValue, 0)
	return nil
}

func (enc *ProtobufEncoder) encode(iValue reflect.Value, flag int) {
	switch iValue.Kind() {
	case reflect.Ptr, reflect.Interface:
		enc.encode(iValue.Elem(), flag)
	case reflect.Struct:
		enc.encodeFlag(flag, 2)
		if flag != 0 {
			enc.encodeVarint(uint64(enc.sizeStruct(iValue)))
		}
		enc.encodeStruct(iValue)
	case reflect.Slice, reflect.Array:
		enc.encodeSlice(iValue, flag)
	case reflect.Map:
		enc.encodeMap(iValue, flag)
	case reflect.String:
		enc.encodeFlag(flag, 2)
		if flag != 0 {
			enc.encodeVarint(uint64(iValue.Len()))
		}
		enc.Write([]byte(iValue.String()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		enc.encodeFlag(flag, 0)
		enc.encodeVarint(uint64(iValue.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		enc.encodeFlag(flag, 0)
		enc.encodeVarint(iValue.Uint())
	case reflect.Float64:
		i := math.Float64bits(iValue.Float())
		enc.encodeFlag(flag, 1)
		enc.Write([]byte{byte(i >> 0), byte(i >> 8), byte(i >> 16), byte(i >> 24), byte(i >> 32), byte(i >> 40), byte(i >> 48), byte(i >> 56)})
	case reflect.Float32:
		i := math.Float32bits(float32(iValue.Float()))
		enc.encodeFlag(flag, 5)
		enc.Write([]byte{byte(i >> 0), byte(i >> 8), byte(i >> 16), byte(i >> 24)})
	case reflect.Bool:
		enc.encodeFlag(flag, 0)
		if iValue.Bool() {
			enc.Write([]byte{1})
			return
		}
		enc.Write([]byte{0})
	}
}

func (enc ProtobufEncoder) encodeStruct(iValue reflect.Value) {
	if iValue.Type() == typeTimeTime {
		t := iValue.Interface().(time.Time)
		enc.Write([]byte{8})
		enc.encodeVarint(uint64(t.Unix()))
		if t.Nanosecond() != 0 {
			enc.Write([]byte{16})
			enc.encodeVarint(uint64(t.Nanosecond()))
		}
		return
	}

	for _, f := range enc.getStructFields(iValue.Type()) {
		field := iValue.Field(f.Index)
		if field.CanInterface() && !field.IsZero() {
			enc.encode(field, f.Offset<<3)
		}
	}
}

func (enc *ProtobufEncoder) getStructFields(iType reflect.Type) []protobufStructField {
	fields, ok := enc.structFields[iType]
	if ok {
		return fields
	}
	for i := 0; i < iType.NumField(); i++ {
		field := newProtobufStructField(i, iType.Field(i))
		if field.Offset != 0 {
			fields = append(fields, field)
		}
	}
	sort.Slice(fields, func(i int, j int) bool {
		return fields[i].Offset < fields[j].Offset
	})
	enc.structFields[iType] = fields
	return fields
}

func (enc ProtobufEncoder) encodeSlice(iValue reflect.Value, flag int) {
	switch iValue.Type().Elem().Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		enc.encodeFlag(flag, 2)
		enc.encodeVarint(enc.sizeMessage(iValue, 0))
		for i := 0; i < iValue.Len(); i++ {
			enc.encode(iValue.Index(i), 0)
		}
	case reflect.Uint8:
		enc.encodeFlag(flag, 2)
		enc.encodeVarint(uint64(iValue.Len()))
		enc.Write(iValue.Bytes())
	default:
		for i := 0; i < iValue.Len(); i++ {
			enc.encode(iValue.Index(i), flag)
		}
	}
}

func (enc ProtobufEncoder) encodeMap(iValue reflect.Value, flag int) {
	for _, key := range iValue.MapKeys() {
		size1 := enc.sizeMessage(key, 8)
		size2 := enc.sizeMessage(iValue.MapIndex(key), 16)
		if size1 == 0 || size2 == 0 {
			return
		}
		break
	}

	for _, key := range iValue.MapKeys() {
		enc.encodeVarint(uint64(flag | 2))
		enc.encodeVarint(enc.sizeMessage(key, 8) + enc.sizeMessage(iValue.MapIndex(key), 16))
		enc.encode(key, 8)
		enc.encode(iValue.MapIndex(key), 16)
	}
}

func (enc ProtobufEncoder) encodeFlag(flag, t int) {
	if flag != 0 {
		enc.encodeVarint(uint64(flag | t))
	}
}
func (enc ProtobufEncoder) encodeVarint(i uint64) {
	n := binary.PutUvarint(enc.varintBuffer[:], i)
	enc.Write(enc.varintBuffer[:n])
}

// SizeMessage 方法计算一个对象编码后的长度。
func (enc *ProtobufEncoder) SizeMessage(i interface{}) uint64 {
	return enc.sizeMessage(reflect.ValueOf(i), 0)
}

func (enc ProtobufEncoder) sizeMessage(iValue reflect.Value, flag int) (size uint64) {
	var msg bool
	switch iValue.Kind() {
	case reflect.Struct:
		msg = true
		size = enc.sizeStruct(iValue)
	case reflect.String:
		msg = true
		size = uint64(iValue.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		size = enc.sizeVarint(uint64(iValue.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		size = enc.sizeVarint(iValue.Uint())
	case reflect.Float64:
		size = 8
	case reflect.Float32:
		size = 4
	case reflect.Bool:
		size = 1
	case reflect.Slice, reflect.Array:
		return enc.sizeSlice(iValue, flag)
	case reflect.Map:
		return enc.sizeMap(iValue, flag)
	case reflect.Ptr, reflect.Interface:
		return enc.sizeMessage(iValue.Elem(), flag)
	}

	if size != 0 && flag != 0 {
		if msg {
			size += enc.sizeVarint(uint64(size))
		}
		size += enc.sizeVarint(uint64(flag))
	}
	return
}

func (enc ProtobufEncoder) sizeSlice(iValue reflect.Value, flag int) uint64 {
	var msg bool
	var size uint64
	switch iValue.Type().Elem().Kind() {
	case reflect.Bool, reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64, reflect.Float32, reflect.Float64:
		msg = true
		for i := 0; i < iValue.Len(); i++ {
			size += enc.sizeMessage(iValue.Index(i), 0)
		}
	case reflect.Uint8:
		msg = true
		size = uint64(iValue.Len())
	default:
		for i := 0; i < iValue.Len(); i++ {
			size += enc.sizeMessage(iValue.Index(i), flag)
		}
		return size
	}
	if size != 0 && flag != 0 {
		if msg {
			size += enc.sizeVarint(uint64(size))
		}
		size += enc.sizeVarint(uint64(flag))
	}
	return size
}

func (enc ProtobufEncoder) sizeMap(iValue reflect.Value, flag int) uint64 {
	var size uint64
	for _, key := range iValue.MapKeys() {
		size1 := enc.sizeMessage(key, 8)
		size2 := enc.sizeMessage(iValue.MapIndex(key), 16)
		if size1 == 0 || size2 == 0 {
			return 0
		}
		size += enc.sizeVarint(uint64(flag)) + enc.sizeVarint(size1+size) + size1 + size2
	}
	return size
}

func (enc ProtobufEncoder) sizeStruct(iValue reflect.Value) uint64 {
	if iValue.Type() == typeTimeTime {
		t := iValue.Interface().(time.Time)
		if t.Nanosecond() == 0 {
			return enc.sizeVarint(uint64(t.Unix())) + 1
		}
		return enc.sizeVarint(uint64(t.Unix())) + enc.sizeVarint(uint64(t.Nanosecond())) + 2
	}

	var size uint64
	for _, f := range enc.getStructFields(iValue.Type()) {
		field := iValue.Field(f.Index)
		if field.CanInterface() && !field.IsZero() {
			size += enc.sizeMessage(field, f.Offset<<3)
		}
	}
	return size
}

func (enc ProtobufEncoder) sizeVarint(i uint64) uint64 {
	return uint64(binary.PutUvarint(enc.varintBuffer[:], i))
}

// ProtobufDecoder 定义protobuf解码器
type ProtobufDecoder struct {
	Reader       *bufio.Reader
	N            int
	Err          error
	LastSlice    map[reflect.Value]int
	varintBuffer [10]byte
}

// NewProtobufDecoder 方法创建一个protobuf解码器。
func NewProtobufDecoder(r io.Reader) *ProtobufDecoder {
	return &ProtobufDecoder{
		Reader: bufio.NewReader(r),
		N:      0xffffffff,
	}
}

// Decode 方法执行解码操作，如果存在数据异常会直接中止解析,会从结构体和tag中获取对应的proto信息。
func (dec *ProtobufDecoder) Decode(i interface{}) error {
	if reflect.Indirect(reflect.ValueOf(i)).Kind() != reflect.Struct {
		return fmt.Errorf(ErrFormatProtobufTypeMustSturct, reflect.TypeOf(i).String())
	}
	dec.decodeValue(reflect.ValueOf(i))
	if dec.Err == io.EOF && dec.N > 0xf0000000 {
		dec.Err = nil
	}
	return dec.Err
}

func (dec *ProtobufDecoder) decodeFlag(iValue reflect.Value, flag uint64) {
	flag = flag & 7
	switch iValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Bool:
		if flag == 0 {
			dec.decodeValue(iValue)
			return
		}
	case reflect.String, reflect.Struct, reflect.Slice, reflect.Map, reflect.Array:
		if flag == 2 {
			dec.decodeMessage(iValue)
			return
		}
	case reflect.Float32:
		if flag == 5 {
			dec.decodeValue(iValue)
			return
		}
	case reflect.Float64:
		if flag == 1 {
			dec.decodeValue(iValue)
			return
		}
	case reflect.Ptr:
		if iValue.IsNil() {
			iValue.Set(reflect.New(iValue.Type().Elem()))
		}
		dec.decodeFlag(iValue.Elem(), flag)
		return
	case reflect.Interface:
		if iValue.IsNil() {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeNilInteface, "flag", iValue.Type())
			return
		}
		dec.decodeFlag(iValue.Elem(), flag)
		return
	}
	dec.Err = fmt.Errorf(ErrFormatProtobufDecodeInvalidKind, "flag", iValue.Kind().String())
}

func (dec *ProtobufDecoder) decodeValue(iValue reflect.Value) {
	switch iValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		iValue.SetInt(int64(dec.decodeVarint()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		iValue.SetUint(dec.decodeVarint())
	case reflect.Bool:
		iValue.SetBool(dec.decodeVarint() == 1)
	case reflect.Float64:
		data := dec.decodeLength(8)
		iValue.SetFloat(math.Float64frombits(uint64(data[7])<<56 | uint64(data[6])<<48 |
			uint64(data[5])<<40 | uint64(data[4])<<32 | uint64(data[3])<<24 |
			uint64(data[2])<<16 | uint64(data[1])<<8 | uint64(data[0])))
	case reflect.Float32:
		data := dec.decodeLength(4)
		iValue.SetFloat(float64(math.Float32frombits(uint32(data[3])<<24 | uint32(data[2])<<16 |
			uint32(data[1])<<8 | uint32(data[0]))))
	case reflect.Struct:
		dec.decodeStruct(iValue)
	case reflect.Slice, reflect.Array:
		dec.decodeSlice(iValue)
	case reflect.Map:
		dec.decodeMap(iValue)
	case reflect.String:
		iValue.SetString(string(dec.decodeLength(dec.N)))
	case reflect.Ptr:
		if iValue.IsNil() {
			iValue.Set(reflect.New(iValue.Type().Elem()))
		}
		dec.decodeValue(iValue.Elem())
	case reflect.Interface:
		if iValue.IsNil() {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeNilInteface, iValue.Type())
			return
		}
		dec.decodeValue(iValue.Elem())
	default:
		dec.Err = fmt.Errorf(ErrFormatProtobufDecodeInvalidKind, "value", iValue.Kind().String())
	}
}

func (dec *ProtobufDecoder) decodeMessage(iValue reflect.Value) {
	pos, size := dec.N, int(dec.decodeVarint())
	if dec.Err != nil {
		dec.Err = fmt.Errorf(ErrFormatProtobufDecodeReadError, "message", "len", dec.Err)
		return
	}
	if size > dec.N || size < 1 {
		dec.Err = fmt.Errorf(ErrFormatProtobufDecodeReadInvalid, "message", size, dec.N)
		return
	}

	dec.N = size
	dec.decodeValue(iValue)
	if dec.N == 0 {
		dec.N = pos - size
	} else if dec.Err == io.EOF {
		dec.Err = fmt.Errorf(ErrFormatProtobufDecodeMessageNotRead, dec.N)
	}
}

func (dec *ProtobufDecoder) decodeSlice(iValue reflect.Value) {
	eType := iValue.Type().Elem()
	switch eType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Bool, reflect.Float64, reflect.Float32:
		for dec.N > 0 && dec.Err == nil {
			val := reflect.New(eType).Elem()
			dec.decodeValue(val)
			iValue.Set(reflect.Append(iValue, val))
		}
	case reflect.Uint8:
		data := dec.decodeLength(dec.N)
		if data != nil {
			iValue.SetBytes(data)
		}
	case reflect.Interface:
		index := dec.getSliceIndex(iValue)
		if index >= iValue.Len() {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeNilInteface, "slice", eType)
			return
		}
		dec.decodeValue(iValue.Index(index))
	default:
		index := dec.getSliceIndex(iValue)
		if index < iValue.Len() {
			dec.decodeValue(iValue.Index(index))
		} else if iValue.Kind() == reflect.Slice {
			val := reflect.New(eType).Elem()
			dec.decodeValue(val)
			iValue.Set(reflect.Append(iValue, val))
		}
	}
}

func (dec *ProtobufDecoder) getSliceIndex(iValue reflect.Value) int {
	if dec.LastSlice == nil {
		dec.LastSlice = make(map[reflect.Value]int)
	}
	index, ok := dec.LastSlice[iValue]
	index++
	if !ok {
		index = 0
	}
	dec.LastSlice[iValue] = index
	return index
}

func (dec *ProtobufDecoder) decodeStruct(iValue reflect.Value) {
	if iValue.Type() == typeTimeTime {
		dec.decodeStructTime(iValue)
		return
	}

	iType := iValue.Type()
	fields := make(map[int]int, iType.NumField())
	for i := 0; i < iType.NumField(); i++ {
		field := newProtobufStructField(i, iType.Field(i))
		if field.Offset != 0 {
			fields[field.Offset] = i
		}
	}
	for dec.N > 0 && dec.Err == nil {
		flag := dec.decodeVarint()
		if dec.Err != nil {
			return
		}

		if flag>>3 == 0 {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeInvalidFlag, "struct", flag)
			return
		}
		index, ok := fields[int(flag>>3)]
		if ok && iValue.Field(index).CanSet() {
			dec.decodeFlag(iValue.Field(index), flag)
		} else {
			dec.decodeStructDiscard(flag)
		}
	}
}

func (dec *ProtobufDecoder) decodeStructTime(iValue reflect.Value) {
	var t1, t2 int64
	for dec.N > 0 && dec.Err == nil {
		flag := dec.decodeVarint()
		if dec.Err != nil {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeReadError, "time", "flag", dec.Err)
			return
		}
		if flag != 8 && flag != 16 {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeInvalidFlag, "time", flag)
			return
		}

		val := int64(dec.decodeVarint())
		if dec.Err != nil {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeReadError, "time", "message", dec.Err)
			return
		}
		if flag == 8 {
			t1 = val
		} else {
			t2 = val
		}
	}
	iValue.Set(reflect.ValueOf(time.Unix(t1, t2)))
}

func (dec *ProtobufDecoder) decodeStructDiscard(flag uint64) {
	switch flag & 7 {
	case 0:
		dec.decodeVarint()
		if dec.Err != nil {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeReadError, "discard", "varint", dec.Err)
		}
	case 1:
		dec.decodeLength(8)
		if dec.Err != nil {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeReadError, "discard", "float64", dec.Err)
		}
	case 2:
		size := dec.decodeVarint()
		if dec.Err != nil {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeReadError, "discard", "message length", dec.Err)
			return
		}
		if size > uint64(dec.N) {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeReadInvalid, "discard message", size, dec.N)
			return
		}
		dec.decodeLength(int(size))
		if dec.Err != nil {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeReadError, "discard", "message", dec.Err)
		}
	case 5:
		dec.decodeLength(4)
		if dec.Err != nil {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeReadError, "discard", "Float32", dec.Err)
		}
	default:
		dec.Err = fmt.Errorf(ErrFormatProtobufDecodeInvalidFlag, "struct discard", flag)
	}
}

func (dec *ProtobufDecoder) decodeMap(iValue reflect.Value) {
	iType := iValue.Type()
	if iValue.IsNil() {
		iValue.Set(reflect.MakeMap(iType))
	}

	var last reflect.Value
	for dec.N > 0 && dec.Err == nil {
		flag := dec.decodeVarint()
		if dec.Err != nil {
			return
		}
		index := flag &^ 7
		if index != 8 && index != 16 {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeInvalidFlag, "map", flag)
			return
		}

		var val reflect.Value
		if index == 8 {
			val = reflect.New(iType.Key())
		} else {
			val = reflect.New(iType.Elem())
		}
		dec.decodeFlag(val, flag)
		if dec.Err != nil {
			dec.Err = fmt.Errorf(ErrFormatProtobufDecodeReadError, "map", "key/value", dec.Err)
			return
		}

		if index == 8 {
			last = val
		} else {
			iValue.SetMapIndex(last.Elem(), val.Elem())
		}
	}
}

func (dec *ProtobufDecoder) decodeLength(length int) []byte {
	buf := make([]byte, length)
	n, err := dec.Reader.Read(buf)
	if err != nil {
		dec.Err = err
		return nil
	}

	dec.N -= n
	if n != length {
		dec.Err = fmt.Errorf(ErrFormatProtobufDecodeReadInvalid, "bytes", length, n)
		return nil
	}
	return buf
}

func (dec *ProtobufDecoder) decodeVarint() uint64 {
	n, err := binary.ReadUvarint(dec.Reader)
	if err != nil {
		dec.Err = err
		return 0
	}

	dec.N -= binary.PutUvarint(dec.varintBuffer[:], n)
	return n
}

type protobufStructField struct {
	Index  int
	Offset int
	Name   string
}

func newProtobufStructField(i int, field reflect.StructField) protobufStructField {
	f := protobufStructField{Index: i, Offset: i + 1, Name: field.Name}
	if field.Tag.Get("protobuf") == "-" {
		return protobufStructField{}
	}
	switch field.Type.Kind() {
	case reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return protobufStructField{}
	}

	for _, tag := range strings.Split(field.Tag.Get("protobuf"), ",") {
		if strings.HasPrefix(tag, "name=") {
			f.Name = tag[5:]
		} else if len(tag) > 0 && 0x2f < tag[0] && tag[0] < 0x3a {
			i, err := strconv.Atoi(tag)
			if err == nil {
				f.Offset = i
			}
		}
	}
	return f
}
