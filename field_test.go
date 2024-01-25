package field

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

func TestJsonRawRawMessageField(t *testing.T) {
	t.Run("newField", func(t *testing.T) {
		if data := NewJSONContent([]byte("{[]}")).Data(); !reflect.DeepEqual(data, json.RawMessage("{[]}")) {
			t.Errorf("invalid newAnyField response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := JsonRawMessage("case1", json.RawMessage("{[]}")).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal any Field: %w", err))
			return
		}
		if result := buf.String(); result != `"case1":{[]}` {
			t.Errorf("invalid marshal JsonRawMessage result: %v", result)
			return
		}
	})
}

func TestFieldList(t *testing.T) {
	var list Fields
	list = append(list, Bool("fieldBool", true))
	list = append(list, Strings("fieldStrings", []string{"aaa", "bbb"}))
	var expected = fmt.Sprintf(`{"fieldBool":true,"fieldStrings":["aaa","bbb"]}`)
	if jBytes, err := list.MarshalJSON(); err != nil {
		t.Error(err)
	} else if !bytes.Equal(jBytes, []byte(expected)) {
		t.Errorf("invalid marshal stringer result: `%s`, expected: `%s`", jBytes, expected)
		return
	}
}

func TestArrayFieldElem(t *testing.T) {
	t.Run("emptyArray", func(t *testing.T) {
		var buf bytes.Buffer
		var _field = ArrayContent{}
		if _data := _field.Data(); reflect.DeepEqual(_data, []Content{}) {
			t.Errorf("invalid array data")
			return
		}
		if err := _field.EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal array Field: %w", err))
			return
		}

		if result := buf.String(); result != `[]` {
			t.Errorf("invalid marshal array result: %v", result)
			return
		}
	})
	t.Run("dataArray", func(t *testing.T) {
		var buf bytes.Buffer
		var _fieldData = []Content{
			NewJSONContent([]byte(`"test1"`)),
			NewJSONContent([]byte(`true`)),
			NewJSONContent([]byte(`undefined`)),
		}
		var _field = ArrayContent{arrayRaw: _fieldData}
		if _data := _field.Data(); reflect.DeepEqual(_data, _fieldData) {
			t.Errorf("invalid array data")
			return
		}
		if err := _field.EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal array Field: %w", err))
			return
		}
		if result := buf.String(); result != `["test1",true,undefined]` {
			t.Errorf("invalid marshal array result: %v", result)
			return
		}
	})
}

func TestBinaryField(t *testing.T) {
	t.Run("newField", func(t *testing.T) {
		if data := NewBinaryContent([]byte{0x12, 0x34}).Data(); !reflect.DeepEqual(data, []byte{0x12, 0x34}) {
			t.Errorf("invalid NewBinaryContent response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Binary("case1", []byte{0x12, 0x34}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal binary Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Binarys("case2", [][]byte{{0x12, 0x34}, {0x56, 0x78}}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal binary Field: %w", err))
			return
		}
		if result := buf.String(); result != `"case1":"data:;base64,EjQ","case2":["data:;base64,EjQ","data:;base64,Vng"]` {
			t.Errorf("invalid marshal binary result: %v", result)
			return
		}
	})
}

func TestBoolField(t *testing.T) {
	t.Run("newField", func(t *testing.T) {
		if data := NewBoolField(true).Data(); data != true {
			t.Errorf("invalid NewBoolField response: %v", data)
			return
		}
		if data := NewBoolField(false).Data(); data != false {
			t.Errorf("invalid NewBoolField response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Bool("case1", true).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal bool Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Bools("case2", []bool{false, true}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal bool Field: %w", err))
			return
		}
		if result := buf.String(); result != `"case1":true,"case2":[false,true]` {
			t.Errorf("invalid marshal bool result: %v", result)
			return
		}
	})
}

func TestComplex64Field(t *testing.T) {
	t.Run("newField", func(t *testing.T) {
		if data := NewComplex64Content(complex(3, -5)).Data(); data != complex64(complex(3, -5)) {
			t.Errorf("invalid NewComplex64Content response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Complex64("case1", complex(3, -5)).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal complex64 Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Complex64s("case2", []complex64{complex(3, -5), complex(4, 6)}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal complex64 Field: %w", err))
			return
		}
		if result := buf.String(); result != `"case1":"(3-5i)","case2":["(3-5i)","(4+6i)"]` {
			t.Errorf("invalid marshal complex64 result: %v", result)
			return
		}
	})
}

func TestComplex128Field(t *testing.T) {
	t.Run("newField", func(t *testing.T) {
		if data := NewComplex128Content(complex(3, -5)).Data(); data != complex(3, -5) {
			t.Errorf("invalid NewComplex128Content response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Complex128("case1", complex(3, -5)).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal complex128 Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Complex128s("case2", []complex128{complex(3, -5), complex(4, 6)}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal complex64 Field: %w", err))
			return
		}
		if result := buf.String(); result != `"case1":"(3-5i)","case2":["(3-5i)","(4+6i)"]` {
			t.Errorf("invalid marshal complex128 result: %v", result)
			return
		}
	})
}

func TestErrorField(t *testing.T) {
	var testErr = fmt.Errorf("test error")
	t.Run("newField", func(t *testing.T) {
		if data := NewErrorContent(testErr).Data(); data != testErr {
			t.Errorf("invalid NewErrorContent response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Error("case1", testErr).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal error Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Errors("case2", []error{testErr, nil}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal error Field: %w", err))
			return
		}
		if result := buf.String(); result != `"case1":"test error","case2":["test error",null]` {
			t.Errorf("invalid marshal error result: %v", result)
			return
		}
	})
}

func TestFloat32Field(t *testing.T) {
	t.Run("newField", func(t *testing.T) {
		if data := NewFloat32Content(1.5).Data(); data != float32(1.5) {
			t.Errorf("invalid NewFloat32Content response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Float32("case1", 1.5).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal float32 Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Float32s("case2", []float32{3.9, 8.15}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal float32 Field: %w", err))
			return
		}
		if result := buf.String(); result != `"case1":1.5,"case2":[3.9,8.15]` {
			t.Errorf("invalid marshal float32 result: %v", result)
			return
		}
	})
}

func TestFloat64Field(t *testing.T) {
	t.Run("newField", func(t *testing.T) {
		if data := NewFloat64Content(1.5).Data(); data != float64(1.5) {
			t.Errorf("invalid NewFloat64Content response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Float64("case1", 1.5).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal float64 Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Float64s("case2", []float64{3.9, 8.15}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal float64 Field: %w", err))
			return
		}
		if result := buf.String(); result != `"case1":1.5,"case2":[3.9,8.15]` {
			t.Errorf("invalid marshal float64 result: %v", result)
			return
		}
	})
}

func TestInt64Field(t *testing.T) {
	var int64A, int64B = rand.Int63(), rand.Int63()
	t.Run("newField", func(t *testing.T) {
		if data := NewIntContent[int64](int64A).Data(); data != int64A {
			t.Errorf("invalid newInt64Field response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Int64("case1", int64A).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal int64 Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Int64s("case2", []int64{int64A, int64B}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal int64 Field: %w", err))
			return
		}
		if result := buf.String(); result != fmt.Sprintf(`"case1":%d,"case2":[%d,%d]`, int64A, int64A, int64B) {
			t.Errorf("invalid marshal int64 result: %v", result)
			return
		}
	})
}

func TestUint64Field(t *testing.T) {
	var uint64A, uint64B = rand.Uint64(), rand.Uint64()
	t.Run("newField", func(t *testing.T) {
		if data := NewUintContent[uint64](uint64A).Data(); data != uint64A {
			t.Errorf("invalid newUint64Field response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Uint64("case1", uint64A).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal uint64 Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Uint64s("case2", []uint64{uint64A, uint64B}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal uint64 Field: %w", err))
			return
		}
		if result := buf.String(); result != fmt.Sprintf(`"case1":%d,"case2":[%d,%d]`, uint64A, uint64A, uint64B) {
			t.Errorf("invalid marshal uint64 result: %v", result)
			return
		}
	})
}

func TestUintptrField(t *testing.T) {
	var uintptrA, uintptrB = uintptr(1), uintptr(2)
	t.Run("newField", func(t *testing.T) {
		if data := NewUintptrContent(uintptrA).Data(); data != uintptrA {
			t.Errorf("invalid NewUintptrContent response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Uintptr("case1", uintptrA).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal uintptr Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Uintptrs("case2", []uintptr{uintptrA, uintptrB}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal uintptr Field: %w", err))
			return
		}
		var expected = fmt.Sprintf(
			`"case1":%#0*x,"case2":[%#0*x,%#0*x]`,
			2*unsafe.Sizeof(uintptrA), uintptrA,
			2*unsafe.Sizeof(uintptrA), uintptrA,
			2*unsafe.Sizeof(uintptrB), uintptrB,
		)
		if result := buf.String(); result != expected {
			t.Errorf("invalid marshal uintptr result: %v, expected: %s", result, expected)
			return
		}
	})
}

func TestStringField(t *testing.T) {
	var stringA, stringB = "t0001", "t0002测试"
	t.Run("newField", func(t *testing.T) {
		if data := NewStringContent(stringA).Data(); data != stringA {
			t.Errorf("invalid NewStringContent response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := String("case1", stringA).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal string Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Strings("case2", []string{stringA, stringB}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal string Field: %w", err))
			return
		}
		var expected = fmt.Sprintf(`"case1":%q,"case2":[%q,%q]`, stringA, stringA, stringB)
		if result := buf.String(); result != expected {
			t.Errorf("invalid marshal string result: `%v`, expected: `%s`", result, expected)
			return
		}
	})
}

type testStringer struct {
	val string
}

func (s *testStringer) String() string { return s.val }

func TestStringerField(t *testing.T) {
	var stringerA, stringerB = &testStringer{val: "t0001"}, &testStringer{val: "t0002"}
	t.Run("newField", func(t *testing.T) {
		if data := NewStringerContent(stringerA).Data(); fmt.Sprint(data) != fmt.Sprint(stringerA) {
			t.Errorf("invalid NewStringerContent response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Stringer("case1", stringerA).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal stringer Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Stringers("case2", []fmt.Stringer{stringerA, stringerB}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal stringer Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Stringer("case3", nil).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal stringer Field: %w", err))
			return
		}
		buf.WriteByte(',')
		var testPtr *testStringer
		if err := Stringer("case4", testPtr).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal stringer Field: %w", err))
			return
		}
		var expected = fmt.Sprintf(`"case1":%q,"case2":[%q,%q],"case3":null,"case4":"<nil>"`, stringerA.val, stringerA.val, stringerB.val)
		if result := buf.String(); result != expected {
			t.Errorf("invalid marshal stringer result: `%v`, expected: `%s`", result, expected)
			return
		}
	})
}

func TestByteStringField(t *testing.T) {
	var byteStringA, byteStringB = []byte("t0001"), []byte("t0002")
	t.Run("newField", func(t *testing.T) {
		if data := NewByteStringContent(byteStringA).Data(); !reflect.DeepEqual(data, string(byteStringA)) {
			t.Errorf("invalid NewByteStringContent response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := ByteString("case1", byteStringA).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal byteString Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := ByteStrings("case2", [][]byte{byteStringA, byteStringB}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal byteString Field: %w", err))
			return
		}
		var expected = fmt.Sprintf(`"case1":%q,"case2":[%q,%q]`, byteStringA, byteStringA, byteStringB)
		if result := buf.String(); result != expected {
			t.Errorf("invalid marshal byteString result: `%v`, expected: `%s`", result, expected)
			return
		}
	})
}

func TestTimeField(t *testing.T) {
	var timezoneB = time.FixedZone("TTT", 60*60*7)
	var timeA, timeB = time.Now(), time.Date(1111, 5, 20, 23, 15, 16, 999111, timezoneB)
	t.Run("newField", func(t *testing.T) {
		if data := NewTimeContent(timeA).Data(); data != timeA {
			t.Errorf("invalid NewTimeContent response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Time("case1", timeA).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal stringer Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Times("case2", []time.Time{timeA, timeB}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal stringer Field: %w", err))
			return
		}
		var expected = fmt.Sprintf(`"case1":%q,"case2":[%q,%q]`, timeA.Format(time.RFC3339), timeA.Format(time.RFC3339), timeB.Format(time.RFC3339))
		if result := buf.String(); result != expected {
			t.Errorf("invalid marshal stringer result: `%v`, expected: `%s`", result, expected)
			return
		}
	})
}

func TestDurationField(t *testing.T) {
	var durationA, durationB = time.Hour - time.Minute*16, time.Minute - time.Second*5
	t.Run("newField", func(t *testing.T) {
		if data := NewStringerContent(durationA).Data(); data != durationA {
			t.Errorf("invalid newDurationField response: %v", data)
			return
		}
	})
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		if err := Duration("case1", durationA).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal stringer Field: %w", err))
			return
		}
		buf.WriteByte(',')
		if err := Durations("case2", []time.Duration{durationA, durationB}).EncodeJSON(&buf); err != nil {
			t.Error(fmt.Errorf("cant marshal stringer Field: %w", err))
			return
		}
		var expected = fmt.Sprintf(`"case1":%q,"case2":[%q,%q]`, durationA, durationA, durationB)
		if result := buf.String(); result != expected {
			t.Errorf("invalid marshal stringer result: `%v`, expected: `%s`", result, expected)
			return
		}
	})
}

type testJSONMarshaler struct{}

func (m testJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(`{"testMarshaler": "123456"}`), nil
}

type testTextMarshaler struct{}

func (m testTextMarshaler) MarshalText() ([]byte, error) {
	return []byte("testMarshaler"), nil
}

type testBinaryMarshaler struct{}

func (m testBinaryMarshaler) MarshalBinary() ([]byte, error) {
	return []byte{0x12, 0x34, 0x56}, nil
}

type funcContent struct{ fn any }

func (f funcContent) Type() Type { return TypeAny }

func (f funcContent) Data() any { return f.fn }

func (f funcContent) EncodeJSON(buffer Buffer) error {
	return errWithoutVal(fmt.Fprintf(buffer, "%q", reflect.TypeOf(f.fn)))
}

type funcTypeInterceptor struct {
}

func (f funcTypeInterceptor) Priority() uint { return 20 }

func (f funcTypeInterceptor) Handle(reflectedType reflect.Type, val any) (Content, bool) {
	if reflectedType.Kind() == reflect.Func {
		return funcContent{fn: val}, true
	}
	return nil, false
}

func TestAnyField(t *testing.T) {
	var (
		boolVal       bool                     = true
		complex128Val complex128               = complex(0, 0)
		complex64Val  complex64                = complex(0, 0)
		durationVal   time.Duration            = time.Second
		float64Val    float64                  = math.MaxFloat64
		float32Val    float32                  = math.MaxFloat32
		intVal        int                      = math.MaxInt
		int64Val      int64                    = math.MaxInt64
		int32Val      int32                    = math.MaxInt32
		int16Val      int16                    = math.MaxInt16
		int8Val       int8                     = math.MaxInt8
		stringVal     string                   = "hello"
		timeVal       time.Time                = time.Unix(100000, 0)
		uintVal       uint                     = math.MaxUint
		uint64Val     uint64                   = math.MaxUint64
		uint32Val     uint32                   = math.MaxUint32
		uint16Val     uint16                   = math.MaxUint16
		uint8Val      uint8                    = math.MaxUint8
		uintptrVal    uintptr                  = 1
		errorVal      error                    = fmt.Errorf("testError")
		stringerVal   fmt.Stringer             = &testStringer{val: "hello world"}
		jsonMarshaler json.Marshaler           = &testJSONMarshaler{}
		textMarshaler encoding.TextMarshaler   = &testTextMarshaler{}
		binMarshaler  encoding.BinaryMarshaler = &testBinaryMarshaler{}
	)

	tests := []struct {
		name   string
		field  Field
		expect Field
	}{
		{"Any:Nil", Any("k", nil), Nil("k")},
		{"Any:Bool", Any("k", true), Bool("k", true)},
		{"Any:Bools", Any("k", []bool{true}), Bools("k", []bool{true})},
		{"Any:Byte", Any("k", byte(1)), Uint8("k", 1)},
		{"Any:Bytes", Any("k", []byte{1}), Binary("k", []byte{1})},
		{"Any:Complex128", Any("k", 1+2i), Complex128("k", 1+2i)},
		{"Any:Complex128s", Any("k", []complex128{1 + 2i}), Complex128s("k", []complex128{1 + 2i})},
		{"Any:Complex64", Any("k", complex64(1+2i)), Complex64("k", 1+2i)},
		{"Any:Complex64s", Any("k", []complex64{1 + 2i}), Complex64s("k", []complex64{1 + 2i})},
		{"Any:Float64", Any("k", 3.14), Float64("k", 3.14)},
		{"Any:Float64s", Any("k", []float64{3.14}), Float64s("k", []float64{3.14})},
		{"Any:Float32", Any("k", float32(3.14)), Float32("k", 3.14)},
		{"Any:Float32s", Any("k", []float32{3.14}), Float32s("k", []float32{3.14})},
		{"Any:Int", Any("k", 1), Int("k", 1)},
		{"Any:Ints", Any("k", []int{1}), Ints("k", []int{1})},
		{"Any:Int64", Any("k", int64(1)), Int64("k", 1)},
		{"Any:Int64s", Any("k", []int64{1}), Int64s("k", []int64{1})},
		{"Any:Int32", Any("k", int32(1)), Int32("k", 1)},
		{"Any:Int32s", Any("k", []int32{1}), Int32s("k", []int32{1})},
		{"Any:Int16", Any("k", int16(1)), Int16("k", 1)},
		{"Any:Int16s", Any("k", []int16{1}), Int16s("k", []int16{1})},
		{"Any:Int8", Any("k", int8(1)), Int8("k", 1)},
		{"Any:Int8s", Any("k", []int8{1}), Int8s("k", []int8{1})},
		{"Any:Rune", Any("k", rune(1)), Int32("k", 1)},
		{"Any:Runes", Any("k", []rune{1}), Int32s("k", []int32{1})},
		{"Any:String", Any("k", "v"), String("k", "v")},
		{"Any:Strings", Any("k", []string{"v"}), Strings("k", []string{"v"})},
		{"Any:Uint", Any("k", uint(1)), Uint("k", uint(1))},
		{"Any:Uints", Any("k", []uint{1}), Uints("k", []uint{1})},
		{"Any:Uint64", Any("k", uint64(1)), Uint64("k", 1)},
		{"Any:Uint64s", Any("k", []uint64{1}), Uint64s("k", []uint64{1})},
		{"Any:Uint32", Any("k", uint32(1)), Uint32("k", 1)},
		{"Any:Uint32s", Any("k", []uint32{1}), Uint32s("k", []uint32{1})},
		{"Any:Uint16", Any("k", uint16(1)), Uint16("k", 1)},
		{"Any:Uint16s", Any("k", []uint16{1}), Uint16s("k", []uint16{1})},
		{"Any:Uint8", Any("k", uint8(1)), Uint8("k", 1)},
		{"Any:Uint8s", Any("k", []uint8{1}), Binary("k", []uint8{1})},
		{"Any:Uintptr", Any("k", uintptr(1)), Uintptr("k", 1)},
		{"Any:Uintptrs", Any("k", []uintptr{1}), Uintptrs("k", []uintptr{1})},
		{"Any:Time", Any("k", time.Unix(0, 0)), Time("k", time.Unix(0, 0))},
		{"Any:Times", Any("k", []time.Time{time.Unix(0, 0)}), Times("k", []time.Time{time.Unix(0, 0)})},
		{"Any:Duration", Any("k", time.Second), Duration("k", time.Second)},
		{"Any:Durations", Any("k", []time.Duration{time.Second}), Durations("k", []time.Duration{time.Second})},
		{"Any:Error", Any("k", errorVal), Error("k", errorVal)},
		{"Any:fmtStringer", Any("k", stringerVal), Stringer("k", stringerVal)},
		{"Any:jsonMarshaler", Any("k", jsonMarshaler), JsonRawMessage("k", valWithoutErr(jsonMarshaler.MarshalJSON()))},
		{"Any:textMarshaler", Any("k", textMarshaler), ByteString("k", valWithoutErr(textMarshaler.MarshalText()))},
		{"Any:binaryMarshaler", Any("k", binMarshaler), Binary("k", valWithoutErr(binMarshaler.MarshalBinary()))},
		{"Any:PtrBool", Any("k", (*bool)(nil)), Nil("k")},
		{"Any:PtrBool", Any("k", &boolVal), Bool("k", boolVal)},
		{"Any:PtrComplex128", Any("k", (*complex128)(nil)), Nil("k")},
		{"Any:PtrComplex128", Any("k", &complex128Val), Complex128("k", complex128Val)},
		{"Any:PtrComplex64", Any("k", (*complex64)(nil)), Nil("k")},
		{"Any:PtrComplex64", Any("k", &complex64Val), Complex64("k", complex64Val)},
		{"Any:PtrDuration", Any("k", (*time.Duration)(nil)), Nil("k")},
		{"Any:PtrDuration", Any("k", &durationVal), Duration("k", durationVal)},
		{"Any:PtrFloat64", Any("k", (*float64)(nil)), Nil("k")},
		{"Any:PtrFloat64", Any("k", &float64Val), Float64("k", float64Val)},
		{"Any:PtrFloat32", Any("k", (*float32)(nil)), Nil("k")},
		{"Any:PtrFloat32", Any("k", &float32Val), Float32("k", float32Val)},
		{"Any:PtrInt", Any("k", (*int)(nil)), Nil("k")},
		{"Any:PtrInt", Any("k", &intVal), Int("k", intVal)},
		{"Any:PtrInt64", Any("k", (*int64)(nil)), Nil("k")},
		{"Any:PtrInt64", Any("k", &int64Val), Int64("k", int64Val)},
		{"Any:PtrInt32", Any("k", (*int32)(nil)), Nil("k")},
		{"Any:PtrInt32", Any("k", &int32Val), Int32("k", int32Val)},
		{"Any:PtrInt16", Any("k", (*int16)(nil)), Nil("k")},
		{"Any:PtrInt16", Any("k", &int16Val), Int16("k", int16Val)},
		{"Any:PtrInt8", Any("k", (*int8)(nil)), Nil("k")},
		{"Any:PtrInt8", Any("k", &int8Val), Int8("k", int8Val)},
		{"Any:PtrString", Any("k", (*string)(nil)), Nil("k")},
		{"Any:PtrString", Any("k", &stringVal), String("k", stringVal)},
		{"Any:PtrTime", Any("k", (*time.Time)(nil)), Nil("k")},
		{"Any:PtrTime", Any("k", &timeVal), Time("k", timeVal)},
		{"Any:PtrUint", Any("k", (*uint)(nil)), Nil("k")},
		{"Any:PtrUint", Any("k", &uintVal), Uint("k", uintVal)},
		{"Any:PtrUint64", Any("k", (*uint64)(nil)), Nil("k")},
		{"Any:PtrUint64", Any("k", &uint64Val), Uint64("k", uint64Val)},
		{"Any:PtrUint32", Any("k", (*uint32)(nil)), Nil("k")},
		{"Any:PtrUint32", Any("k", &uint32Val), Uint32("k", uint32Val)},
		{"Any:PtrUint16", Any("k", (*uint16)(nil)), Nil("k")},
		{"Any:PtrUint16", Any("k", &uint16Val), Uint16("k", uint16Val)},
		{"Any:PtrUint8", Any("k", (*uint8)(nil)), Nil("k")},
		{"Any:PtrUint8", Any("k", &uint8Val), Uint8("k", uint8Val)},
		{"Any:PtrUintptr", Any("k", (*uintptr)(nil)), Nil("k")},
		{"Any:PtrUintptr", Any("k", &uintptrVal), Uintptr("k", uintptrVal)},
		{"Any:PtrError", Any("k", &errorVal), Error("k", errorVal)},
	}

	for _, testItem := range tests {
		t.Run(testItem.name, func(t *testing.T) {
			var buf1, buf2 bytes.Buffer
			if err := testItem.field.EncodeJSON(&buf1); err != nil {
				t.Error(err)
				return
			}
			if err := testItem.expect.EncodeJSON(&buf2); err != nil {
				t.Error(err)
				return
			}
			if b1, b2 := buf1.Bytes(), buf2.Bytes(); !bytes.Equal(b1, b2) {
				t.Error(fmt.Errorf("not equal:\n\texpected: %s\n\tgot:%s", b2, b1))
			}
		})
	}
}
