package field

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type Fields []Field

func (f Fields) Unique() []Field {
	var newFields = make([]Field, 0, len(f))
	var m = make(map[string]struct{}, len(f))
	var newCount = 0
	for i := 0; i < len(f); i++ {
		if _, exist := m[f[i].Key]; !exist {
			m[f[i].Key] = struct{}{}
			newFields = append(newFields, f[i])
			newCount++
		}
	}
	newFields = newFields[:newCount]
	sort.SliceStable(newFields, func(i, j int) bool {
		return strings.Compare(newFields[i].Key, newFields[j].Key) < 0
	})
	return newFields
}

func (f Fields) Has(key string) bool {
	for i := 0; i < len(f); i++ {
		if f[i].Key == key {
			return true
		}
	}
	return false
}

func (f Fields) Get(key string) (Field, bool) {
	for i := 0; i < len(f); i++ {
		if f[i].Key == key {
			return f[i], true
		}
	}
	return Field{}, false
}

func (f Fields) Export() map[string]any {
	var m = make(map[string]any, len(f))
	for i := 0; i < len(f); i++ {
		m[f[i].Key] = f[i].Data()
	}
	return m
}

func (f Fields) EncodeJSON(buf Buffer) (err error) {
	var snap = f.Unique()
	if err = buf.WriteByte('{'); err != nil {
		return err
	}
	for i := 0; i < len(snap); i++ {
		if i > 0 {
			if err = buf.WriteByte(','); err != nil {
				return err
			}
		}
		if err = snap[i].EncodeJSON(buf); err != nil {
			return err
		}
	}
	if err = buf.WriteByte('}'); err != nil {
		return err
	}
	return nil
}

func (f Fields) MarshalJSON() (dst []byte, err error) {
	var buf bytes.Buffer
	err = f.EncodeJSON(&buf)
	return buf.Bytes(), err
}

type Type uint8

const (
	TypeNull = iota
	TypeBool
	TypeInt
	TypeUint
	TypeUintptr
	TypeFloat
	TypeComplex
	TypeString
	TypeStringer
	TypeBinary
	TypeTime
	TypeError
	TypeJSON
	TypeAny
	TypeArray = 0x80
)

func valWithoutErr[T any](val T, _ error) T { return val }

func errWithoutVal[T any](_ T, err error) error { return err }

type Buffer interface {
	Write(p []byte) (n int, err error)
	WriteString(s string) (n int, err error)
	WriteByte(c byte) error
	WriteRune(r rune) (n int, err error)
}

type Content interface {
	Type() Type
	Data() any
	EncodeJSON(buffer Buffer) error
}

type Field struct {
	Key string
	Content
}

func (f Field) EncodeJSON(buffer Buffer) (err error) {
	if err = appendJsonStringBuf(buffer, f.Key); err != nil {
		return err
	}
	if err = buffer.WriteByte(':'); err != nil {
		return err
	}
	return f.Content.EncodeJSON(buffer)
}

func (f Field) MarshalJSON() (_ []byte, err error) {
	var buf bytes.Buffer
	if err = buf.WriteByte('{'); err != nil {
		return nil, err
	}
	if err = f.EncodeJSON(&buf); err != nil {
		return nil, err
	}
	if err = buf.WriteByte('}'); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// array wrapper

func newArray[T any](list []T, converter func(T) Content) ArrayContent {
	var result = make([]Content, len(list))
	for i := 0; i < len(list); i++ {
		result[i] = converter(list[i])
	}
	return ArrayContent{arrayRaw: result}
}

type ArrayContent struct {
	arrayRaw []Content
}

func (f ArrayContent) Type() Type {
	if len(f.arrayRaw) == 0 {
		return TypeArray | TypeNull
	}
	return TypeArray | f.arrayRaw[0].Type()
}

func (f ArrayContent) Data() any {
	var data = make([]any, len(f.arrayRaw))
	for i := 0; i < len(f.arrayRaw); i++ {
		data[i] = f.arrayRaw[i].Data()
	}
	return data
}

func (f ArrayContent) Raw() []Content { return f.arrayRaw }

func (f ArrayContent) EncodeJSON(buffer Buffer) (err error) {
	if err = buffer.WriteByte('['); err != nil {
		return err
	}
	for i := 0; i < len(f.arrayRaw); i++ {
		if i > 0 {
			if err = buffer.WriteByte(','); err != nil {
				return err
			}
		}
		if err = f.arrayRaw[i].EncodeJSON(buffer); err != nil {
			return err
		}
	}
	if err = buffer.WriteByte(']'); err != nil {
		return err
	}
	return nil
}

type NilContent struct{}

func NewNilContent() Content { return NilContent{} }

func (n NilContent) Type() Type { return TypeNull }

func (n NilContent) Data() any { return nil }

func (n NilContent) EncodeJSON(buffer Buffer) error {
	return errWithoutVal(buffer.Write([]byte{'n', 'u', 'l', 'l'}))
}

func Nil(key string) Field { return Field{Key: key, Content: NilContent{}} }

// data type: jsonRawRawMessage

type JSONContent struct{ jsonRaw json.RawMessage }

func NewJSONContent(val []byte) Content { return JSONContent{jsonRaw: val} }

func (f JSONContent) Type() Type { return TypeJSON }

func (f JSONContent) Data() any { return f.jsonRaw }

func (f JSONContent) Raw() json.RawMessage { return f.jsonRaw }

func (f JSONContent) EncodeJSON(buffer Buffer) (err error) {
	return errWithoutVal(buffer.Write(f.jsonRaw))
}

func JsonRawMessage(key string, val json.RawMessage) Field {
	return Field{key, NewJSONContent(val)}
}

// data type: binary

type BinaryContent struct{ binaryRaw []byte }

func NewBinaryContent(val []byte) Content { return BinaryContent{binaryRaw: val} }

func (f BinaryContent) Type() Type { return TypeBinary }

func (f BinaryContent) Data() any { return f.binaryRaw }

func (f BinaryContent) Raw() json.RawMessage { return f.binaryRaw }

func (f BinaryContent) String() string { return base64.StdEncoding.EncodeToString(f.binaryRaw) }

func (f BinaryContent) EncodeJSON(buffer Buffer) (err error) {
	if _, err = buffer.WriteString(`"data:;base64,`); err != nil {
		return err
	}
	var encoder = base64.NewEncoder(base64.RawStdEncoding, buffer)
	if _, err = encoder.Write(f.binaryRaw); err != nil {
		return err
	}
	if err = encoder.Close(); err != nil {
		return err
	}
	if err = buffer.WriteByte('"'); err != nil {
		return err
	}
	return nil
}

func Binary(key string, val []byte) Field { return Field{key, NewBinaryContent(val)} }

func Binarys(key string, valArr [][]byte) Field {
	return Field{key, newArray(valArr, NewBinaryContent)}
}

// data type: bool

type BoolContent bool

func NewBoolField(val bool) Content { return BoolContent(val) }

func (f BoolContent) Type() Type { return TypeBool }

func (f BoolContent) Data() any { return bool(f) }

func (f BoolContent) Raw() bool { return bool(f) }

func (f BoolContent) EncodeJSON(buffer Buffer) error {
	if f {
		return errWithoutVal(buffer.Write([]byte{'t', 'r', 'u', 'e'}))
	} else {
		return errWithoutVal(buffer.Write([]byte{'f', 'a', 'l', 's', 'e'}))
	}
}

func Bool(key string, val bool) Field { return Field{Key: key, Content: NewBoolField(val)} }

func Bools(key string, valArr []bool) Field { return Field{key, newArray(valArr, NewBoolField)} }

// data type: complex128

type Complex128Content complex128

func NewComplex128Content(val complex128) Content { return Complex128Content(val) }

func (f Complex128Content) Type() Type { return TypeComplex }

func (f Complex128Content) Data() any { return complex128(f) }

func (f Complex128Content) Raw() complex128 { return complex128(f) }

func (f Complex128Content) EncodeJSON(buffer Buffer) (err error) {
	return appendJsonStringBuf(buffer, strconv.FormatComplex(complex128(f), 'f', -1, 128))
}

func Complex128(key string, val complex128) Field {
	return Field{Key: key, Content: NewComplex128Content(val)}
}

func Complex128s(key string, nums []complex128) Field {
	return Field{key, newArray(nums, NewComplex128Content)}
}

// data type: complex64

type Complex64Content complex64

func NewComplex64Content(val complex64) Content { return Complex64Content(val) }

func (f Complex64Content) Type() Type { return TypeComplex }

func (f Complex64Content) Data() any { return complex64(f) }

func (f Complex64Content) Raw() complex64 { return complex64(f) }

func (f Complex64Content) EncodeJSON(buffer Buffer) (err error) {
	return appendJsonStringBuf(buffer, strconv.FormatComplex(complex128(f), 'f', -1, 64))
}

func Complex64(key string, val complex64) Field {
	return Field{Key: key, Content: NewComplex64Content(val)}
}

func Complex64s(key string, nums []complex64) Field {
	return Field{key, newArray(nums, NewComplex64Content)}
}

// data type: error

type ErrorContent struct{ data error }

func NewErrorContent(val error) Content { return ErrorContent{data: val} }

func (f ErrorContent) Type() Type { return TypeError }

func (f ErrorContent) Data() any { return f.data }

func (f ErrorContent) Raw() error { return f.data }

func (f ErrorContent) EncodeJSON(buffer Buffer) error {
	if f.data == nil {
		return errWithoutVal(buffer.Write([]byte{'n', 'u', 'l', 'l'}))
	}
	if content, err := asJsonErrMarshaler(f.data).MarshalJSON(); err == nil {
		return errWithoutVal(buffer.Write(content))
	} else {
		return err
	}
}

func Error(key string, err error) Field { return Field{Key: key, Content: NewErrorContent(err)} }

func Errors(key string, errs []error) Field { return Field{key, newArray(errs, NewErrorContent)} }

// data type: float32

type Float32Content float32

func NewFloat32Content(val float32) Content { return Float32Content(val) }

func (f Float32Content) Type() Type { return TypeFloat }

func (f Float32Content) Data() any { return float32(f) }

func (f Float32Content) Raw() float32 { return float32(f) }

func (f Float32Content) EncodeJSON(buffer Buffer) (err error) {
	return errWithoutVal(buffer.WriteString(strconv.FormatFloat(float64(f), 'f', -1, 32)))
}

func Float32(key string, val float32) Field { return Field{Key: key, Content: NewFloat32Content(val)} }

func Float32s(key string, nums []float32) Field { return Field{key, newArray(nums, NewFloat32Content)} }

// data type: float64

type Float64Content float64

func NewFloat64Content(val float64) Content { return Float64Content(val) }

func (f Float64Content) Type() Type { return TypeFloat }

func (f Float64Content) Data() any { return float64(f) }

func (f Float64Content) Raw() float64 { return float64(f) }

func (f Float64Content) EncodeJSON(buffer Buffer) (err error) {
	return errWithoutVal(buffer.WriteString(strconv.FormatFloat(float64(f), 'f', -1, 64)))
}

func Float64(key string, val float64) Field { return Field{Key: key, Content: NewFloat64Content(val)} }

func Float64s(key string, nums []float64) Field { return Field{key, newArray(nums, NewFloat64Content)} }

// data type: int

type IntContent[T int | int8 | int16 | int32 | int64] struct {
	data T
}

func NewIntContent[T int | int8 | int16 | int32 | int64](val T) Content {
	return IntContent[T]{data: val}
}

func (f IntContent[T]) Type() Type { return TypeInt }

func (f IntContent[T]) Data() any { return f.data }

func (f IntContent[T]) Raw() T { return f.data }

func (f IntContent[T]) EncodeJSON(buffer Buffer) (err error) {
	return errWithoutVal(buffer.WriteString(strconv.FormatInt(int64(f.data), 10)))
}

func Int[T int | int8 | int16 | int32 | int64](key string, val T) Field {
	return Field{Key: key, Content: NewIntContent(val)}
}

func Ints[T int | int8 | int16 | int32 | int64](key string, nums []T) Field {
	return Field{key, newArray(nums, NewIntContent[T])}
}

func Int8(key string, val int8) Field {
	return Field{Key: key, Content: NewIntContent(val)}
}

func Int8s(key string, nums []int8) Field {
	return Field{key, newArray(nums, NewIntContent[int8])}
}

func Int16(key string, val int16) Field {
	return Field{Key: key, Content: NewIntContent(val)}
}

func Int16s(key string, nums []int16) Field {
	return Field{key, newArray(nums, NewIntContent[int16])}
}

func Int32(key string, val int32) Field {
	return Field{Key: key, Content: NewIntContent(val)}
}

func Int32s(key string, nums []int32) Field {
	return Field{key, newArray(nums, NewIntContent[int32])}
}

func Int64(key string, val int64) Field {
	return Field{Key: key, Content: NewIntContent(val)}
}

func Int64s(key string, nums []int64) Field {
	return Field{key, newArray(nums, NewIntContent[int64])}
}

// data type: uint

type UintContent[T uint | uint8 | uint16 | uint32 | uint64] struct {
	data T
}

func NewUintContent[T uint | uint8 | uint16 | uint32 | uint64](val T) Content {
	return UintContent[T]{data: val}
}

func (f UintContent[T]) Type() Type { return TypeUint }

func (f UintContent[T]) Data() any { return f.data }

func (f UintContent[T]) Raw() T { return f.data }

func (f UintContent[T]) EncodeJSON(buffer Buffer) (err error) {
	return errWithoutVal(buffer.WriteString(strconv.FormatUint(uint64(f.data), 10)))
}

func Uint[T uint | uint8 | uint16 | uint32 | uint64](key string, val T) Field {
	return Field{Key: key, Content: NewUintContent(val)}
}

func Uints[T uint | uint8 | uint16 | uint32 | uint64](key string, nums []T) Field {
	return Field{key, newArray(nums, NewUintContent[T])}
}

func Uint8(key string, val uint8) Field {
	return Field{Key: key, Content: NewUintContent(val)}
}

func Uint8s(key string, nums []uint8) Field {
	return Field{key, newArray(nums, NewUintContent[uint8])}
}

func Uint16(key string, val uint16) Field {
	return Field{Key: key, Content: NewUintContent(val)}
}

func Uint16s(key string, nums []uint16) Field {
	return Field{key, newArray(nums, NewUintContent[uint16])}
}

func Uint32(key string, val uint32) Field {
	return Field{Key: key, Content: NewUintContent(val)}
}

func Uint32s(key string, nums []uint32) Field {
	return Field{key, newArray(nums, NewUintContent[uint32])}
}

func Uint64(key string, val uint64) Field {
	return Field{Key: key, Content: NewUintContent(val)}
}

func Uint64s(key string, nums []uint64) Field {
	return Field{key, newArray(nums, NewUintContent[uint64])}
}

// data type: uintptr

type UintptrContent uintptr

func NewUintptrContent(val uintptr) Content { return UintptrContent(val) }

func (f UintptrContent) Type() Type { return TypeUintptr }

func (f UintptrContent) Data() any { return uintptr(f) }

func (f UintptrContent) Raw() uintptr { return uintptr(f) }

func (f UintptrContent) EncodeJSON(buffer Buffer) (err error) {
	return errWithoutVal(fmt.Fprintf(buffer, "%#0*x", 2*unsafe.Sizeof(f), f.Data()))
}

func Uintptr(key string, val uintptr) Field {
	return Field{Key: key, Content: UintptrContent(val)}
}

func Uintptrs(key string, us []uintptr) Field {
	return Field{Key: key, Content: newArray(us, NewUintptrContent)}
}

// data type: string

type StringContent string

func NewStringContent(val string) Content { return StringContent(val) }

func (f StringContent) Type() Type { return TypeString }

func (f StringContent) Data() any { return string(f) }

func (f StringContent) Raw() string { return string(f) }

func (f StringContent) EncodeJSON(buffer Buffer) error {
	return appendJsonStringBuf(buffer, string(f))
}

func String(key string, val string) Field {
	return Field{Key: key, Content: NewStringContent(val)}
}

func Strings(key string, valArr []string) Field {
	return Field{Key: key, Content: newArray(valArr, NewStringContent)}
}

// data type: byteString

func NewByteStringContent(val []byte) Content { return StringContent(val) }

func ByteString(key string, val []byte) Field {
	return Field{Key: key, Content: NewStringContent(string(val))}
}

func ByteStrings(key string, valArr [][]byte) Field {
	return Field{Key: key, Content: newArray(valArr, NewByteStringContent)}
}

// data type: stringer

type StringerContent struct{ data fmt.Stringer }

func NewStringerContent[T fmt.Stringer](val T) Content { return StringerContent{data: val} }

func (f StringerContent) Type() Type { return TypeStringer }

func (f StringerContent) Data() any { return f.data }

func (f StringerContent) Raw() fmt.Stringer { return f.data }

func (f StringerContent) EncodeJSON(buffer Buffer) (err error) {
	if f.data == nil {
		return errWithoutVal(buffer.Write([]byte{'n', 'u', 'l', 'l'}))
	}
	return appendJsonStringBuf(buffer, fmt.Sprintf("%s", f.data))
}

func Stringer(key string, val fmt.Stringer) Field {
	return Field{Key: key, Content: NewStringerContent(val)}
}

func Stringers[T fmt.Stringer](key string, valArr []T) Field {
	return Field{Key: key, Content: newArray(valArr, NewStringerContent[T])}
}

// data type: time

type TimeContent time.Time

func NewTimeContent(val time.Time) Content { return TimeContent(val) }

func (f TimeContent) Type() Type { return TypeTime }

func (f TimeContent) Data() any { return time.Time(f) }

func (f TimeContent) Raw() time.Time { return time.Time(f) }

func (f TimeContent) EncodeJSON(buffer Buffer) error {
	return appendJsonStringBuf(buffer, time.Time(f).Format(time.RFC3339))
}

func Time(key string, val time.Time) Field {
	return Field{Key: key, Content: NewTimeContent(val)}
}

func Times(key string, valArr []time.Time) Field {
	return Field{Key: key, Content: newArray(valArr, NewTimeContent)}
}

// data type: duration

func Duration(key string, val time.Duration) Field { return Stringer(key, val) }

func Durations(key string, valArr []time.Duration) Field { return Stringers(key, valArr) }

func anyPointer[T any](key string, ptr *T, fn func(string, T) Field) Field {
	if ptr == nil {
		return Nil(key)
	}
	return fn(key, *ptr)
}

type AnyTypeInterceptor interface {
	Priority() uint
	Handle(reflectedType reflect.Type, val any) (Content, bool)
}

func Any(key string, val any) Field {
	switch v := val.(type) {
	case nil:
		return Nil(key)
	case bool:
		return Bool(key, v)
	case *bool:
		return anyPointer(key, v, Bool)
	case []bool:
		return Bools(key, v)
	case complex128:
		return Complex128(key, v)
	case *complex128:
		return anyPointer(key, v, Complex128)
	case []complex128:
		return Complex128s(key, v)
	case complex64:
		return Complex64(key, v)
	case *complex64:
		return anyPointer(key, v, Complex64)
	case []complex64:
		return Complex64s(key, v)
	case float64:
		return Float64(key, v)
	case *float64:
		return anyPointer(key, v, Float64)
	case []float64:
		return Float64s(key, v)
	case float32:
		return Float32(key, v)
	case *float32:
		return anyPointer(key, v, Float32)
	case []float32:
		return Float32s(key, v)
	case int:
		return Int(key, v)
	case *int:
		return anyPointer(key, v, Int[int])
	case []int:
		return Ints(key, v)
	case int64:
		return Int(key, v)
	case *int64:
		return anyPointer(key, v, Int64)
	case []int64:
		return Int64s(key, v)
	case int32:
		return Int(key, v)
	case *int32:
		return anyPointer(key, v, Int32)
	case []int32:
		return Int32s(key, v)
	case int16:
		return Int(key, v)
	case *int16:
		return anyPointer(key, v, Int16)
	case []int16:
		return Int16s(key, v)
	case int8:
		return Int(key, v)
	case *int8:
		return anyPointer(key, v, Int8)
	case []int8:
		return Int8s(key, v)
	case string:
		return String(key, v)
	case *string:
		return anyPointer(key, v, String)
	case []string:
		return Strings(key, v)
	case uint:
		return Uint(key, v)
	case *uint:
		return anyPointer(key, v, Uint[uint])
	case []uint:
		return Uints(key, v)
	case uint64:
		return Uint(key, v)
	case *uint64:
		return anyPointer(key, v, Uint64)
	case []uint64:
		return Uint64s(key, v)
	case uint32:
		return Uint(key, v)
	case *uint32:
		return anyPointer(key, v, Uint32)
	case []uint32:
		return Uint32s(key, v)
	case uint16:
		return Uint(key, v)
	case *uint16:
		return anyPointer(key, v, Uint16)
	case []uint16:
		return Uint16s(key, v)
	case uint8:
		return Uint(key, v)
	case *uint8:
		return anyPointer(key, v, Uint8)
	case []byte:
		return Binary(key, v)
	case uintptr:
		return Uintptr(key, v)
	case *uintptr:
		return anyPointer(key, v, Uintptr)
	case []uintptr:
		return Uintptrs(key, v)
	case time.Time:
		return Time(key, v)
	case *time.Time:
		return anyPointer(key, v, Time)
	case []time.Time:
		return Times(key, v)
	case time.Duration:
		return Duration(key, v)
	case *time.Duration:
		return anyPointer(key, v, Duration)
	case []time.Duration:
		return Durations(key, v)
	case error:
		return Error(key, v)
	case *error:
		return anyPointer(key, v, Error)
	case []error:
		return Errors(key, v)
	case fmt.Stringer:
		return Stringer(key, v)
	case json.Marshaler:
		if content, err := v.MarshalJSON(); err != nil {
			return Error(key, fmt.Errorf("cant marshal json: %w", err))
		} else {
			return JsonRawMessage(key, content)
		}
	case encoding.TextMarshaler:
		if content, err := v.MarshalText(); err != nil {
			return Error(key, fmt.Errorf("cant marshal text: %w", err))
		} else {
			return ByteString(key, content)
		}
	case encoding.BinaryMarshaler:
		if content, err := v.MarshalBinary(); err != nil {
			return Error(key, fmt.Errorf("cant marshal binary: %w", err))
		} else {
			return Binary(key, content)
		}
	}
	return Error(key, fmt.Errorf("cant marshal field: no type matched"))
}
