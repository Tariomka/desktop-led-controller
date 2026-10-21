package show

import (
	"encoding/binary"
	"errors"
	"math"
	"strings"
	"unicode/utf8"
)

// A reader for the subset of MS-NRBF (.NET BinaryFormatter) used by the vendor 3D Software
// project files. Values are decoded into nrbfObject (classes), []any (object arrays),
// []byte (byte arrays), string, bool and Go integer/float types; references are resolved
// through nrbfStream.value.

var ErrNRBF = errors.New("invalid .NET binary serialization data")

// Record types (MS-NRBF 2.1.2.1).
const (
	recordHeader                = 0
	recordClassWithID           = 1
	recordSystemClassWithMember = 2
	recordClassWithMembers      = 3
	recordSystemClassWithTypes  = 4
	recordClassWithTypes        = 5
	recordObjectString          = 6
	recordBinaryArray           = 7
	recordMemberPrimitiveTyped  = 8
	recordMemberReference       = 9
	recordObjectNull            = 10
	recordMessageEnd            = 11
	recordBinaryLibrary         = 12
	recordObjectNullMultiple256 = 13
	recordObjectNullMultiple    = 14
	recordArraySinglePrimitive  = 15
	recordArraySingleObject     = 16
	recordArraySingleString     = 17
)

// Binary types of class members (MS-NRBF 2.1.2.2).
const (
	binaryPrimitive      = 0
	binaryString         = 1
	binaryObject         = 2
	binarySystemClass    = 3
	binaryClass          = 4
	binaryObjectArray    = 5
	binaryStringArray    = 6
	binaryPrimitiveArray = 7
)

// Primitive types (MS-NRBF 2.1.2.3).
const (
	primitiveBoolean  = 1
	primitiveByte     = 2
	primitiveChar     = 3
	primitiveDecimal  = 5
	primitiveDouble   = 6
	primitiveInt16    = 7
	primitiveInt32    = 8
	primitiveInt64    = 9
	primitiveSByte    = 10
	primitiveSingle   = 11
	primitiveTimeSpan = 12
	primitiveDateTime = 13
	primitiveUInt16   = 14
	primitiveUInt32   = 15
	primitiveUInt64   = 16
	primitiveNull     = 17
	primitiveString   = 18
)

type nrbfClass struct {
	name       string
	members    []string
	types      []byte // binary type per member; nil when the record carries no type info
	primitives []byte // primitive type per member (binaryPrimitive / binaryPrimitiveArray)
}

type nrbfObject struct {
	class  *nrbfClass
	values []any
}

// get returns the raw value of a member; inherited members are named "Base+member".
func (this *nrbfObject) get(name string) any {
	for i, member := range this.class.members {
		if member == name || strings.HasSuffix(member, "+"+name) {
			return this.values[i]
		}
	}
	return nil
}

type nrbfReference int32

type nrbfNulls int // a run of null array elements

// nrbfReader reads consecutive serialization streams from data.
type nrbfReader struct {
	data []byte
	pos  int
	err  error
}

// nrbfStream holds the objects of one stream, so member references can be resolved.
type nrbfStream struct {
	*nrbfReader
	objects map[int32]any
	classes map[int32]*nrbfClass
}

func (this *nrbfReader) fail() {
	if this.err == nil {
		this.err = ErrNRBF
	}
}

func (this *nrbfReader) take(n int) []byte {
	if this.err != nil || n < 0 || n > len(this.data)-this.pos {
		this.fail()
		return make([]byte, max(n, 0))
	}
	result := this.data[this.pos : this.pos+n]
	this.pos += n
	return result
}

func (this *nrbfReader) u8() byte    { return this.take(1)[0] }
func (this *nrbfReader) i32() int32  { return int32(binary.LittleEndian.Uint32(this.take(4))) }
func (this *nrbfReader) u16() uint16 { return binary.LittleEndian.Uint16(this.take(2)) }
func (this *nrbfReader) u32() uint32 { return binary.LittleEndian.Uint32(this.take(4)) }
func (this *nrbfReader) u64() uint64 { return binary.LittleEndian.Uint64(this.take(8)) }

// count reads an element count and checks it against the remaining data.
func (this *nrbfReader) count(minSize int) int {
	n := this.i32()
	if n < 0 || int(n)*minSize > len(this.data)-this.pos {
		this.fail()
		return 0
	}
	return int(n)
}

func (this *nrbfReader) text() string {
	length := 0
	for shift := 0; shift < 35; shift += 7 {
		b := this.u8()
		length |= int(b&0x7F) << shift
		if b&0x80 == 0 {
			return string(this.take(length))
		}
	}
	this.fail()
	return ""
}

// stream reads one serialization stream and returns it with its root object.
func (this *nrbfReader) stream() (*nrbfStream, any) {
	stream := &nrbfStream{nrbfReader: this, objects: map[int32]any{}, classes: map[int32]*nrbfClass{}}
	if this.u8() != recordHeader {
		this.fail()
		return stream, nil
	}
	root := this.i32()
	this.take(12) // header id, major and minor version

	for this.err == nil {
		value, end := stream.record()
		if end {
			break
		}
		if _, ok := value.(nrbfNulls); ok {
			this.fail()
		}
	}
	return stream, stream.value(nrbfReference(root))
}

// record reads one record; end is true for MessageEnd.
func (this *nrbfStream) record() (value any, end bool) {
	kind := this.u8()
	for kind == recordBinaryLibrary && this.err == nil {
		this.i32()
		this.text()
		kind = this.u8()
	}
	if this.err != nil {
		return nil, true
	}

	switch kind {
	case recordClassWithID:
		id, class := this.i32(), this.classes[this.i32()]
		if class == nil {
			this.fail()
			return nil, true
		}
		return this.object(id, class), false
	case recordSystemClassWithMember, recordClassWithMembers, recordSystemClassWithTypes, recordClassWithTypes:
		id, class := this.classInfo(kind)
		return this.object(id, class), false
	case recordObjectString:
		id := this.i32()
		value := this.text()
		this.objects[id] = value
		return value, false
	case recordBinaryArray:
		return this.binaryArray(), false
	case recordMemberPrimitiveTyped:
		return this.primitive(this.u8()), false
	case recordMemberReference:
		return nrbfReference(this.i32()), false
	case recordObjectNull:
		return nil, false
	case recordMessageEnd:
		return nil, true
	case recordObjectNullMultiple256:
		return nrbfNulls(this.u8()), false
	case recordObjectNullMultiple:
		return nrbfNulls(this.count(0)), false
	case recordArraySinglePrimitive:
		id, length := this.i32(), this.count(1)
		value := this.primitiveArray(this.u8(), length)
		this.objects[id] = value
		return value, false
	case recordArraySingleObject, recordArraySingleString:
		id, length := this.i32(), this.count(1)
		value := this.elements(length)
		this.objects[id] = value
		return value, false
	}
	this.fail()
	return nil, true
}

func (this *nrbfStream) classInfo(kind byte) (int32, *nrbfClass) {
	id := this.i32()
	class := &nrbfClass{name: this.text()}
	for range this.count(1) {
		class.members = append(class.members, this.text())
	}

	if kind == recordSystemClassWithTypes || kind == recordClassWithTypes {
		class.types = this.take(len(class.members))
		class.primitives = make([]byte, len(class.members))
		for i, binaryType := range class.types {
			class.primitives[i] = this.typeInfo(binaryType)
		}
	}
	if kind == recordClassWithMembers || kind == recordClassWithTypes {
		this.i32() // library id
	}
	this.classes[id] = class
	return id, class
}

// typeInfo reads the additional type information of a binary type; it returns the primitive
// type for primitives and primitive arrays.
func (this *nrbfStream) typeInfo(binaryType byte) byte {
	switch binaryType {
	case binaryPrimitive, binaryPrimitiveArray:
		return this.u8()
	case binarySystemClass:
		this.text()
	case binaryClass:
		this.text()
		this.i32()
	case binaryString, binaryObject, binaryObjectArray, binaryStringArray:
	default:
		this.fail()
	}
	return 0
}

func (this *nrbfStream) object(id int32, class *nrbfClass) *nrbfObject {
	object := &nrbfObject{class: class, values: make([]any, len(class.members))}
	this.objects[id] = object
	for i := range class.members {
		if class.types != nil && class.types[i] == binaryPrimitive {
			object.values[i] = this.primitive(class.primitives[i])
		} else {
			object.values[i] = this.memberRecord()
		}
		if this.err != nil {
			break
		}
	}
	return object
}

func (this *nrbfStream) memberRecord() any {
	value, end := this.record()
	if _, nulls := value.(nrbfNulls); end || nulls {
		this.fail()
		return nil
	}
	return value
}

func (this *nrbfStream) binaryArray() any {
	id, arrayType := this.i32(), this.u8()
	rank, length := this.count(4), 1
	for range rank {
		length *= this.count(0)
	}
	if arrayType >= 3 && arrayType <= 5 { // offset variants carry lower bounds
		this.take(4 * rank)
	}
	if length > len(this.data)-this.pos {
		this.fail()
		return nil
	}

	binaryType := this.u8()
	primitive := this.typeInfo(binaryType)
	var value any
	if binaryType == binaryPrimitive {
		value = this.primitiveArray(primitive, length)
	} else {
		value = this.elements(length)
	}
	this.objects[id] = value
	return value
}

func (this *nrbfStream) elements(length int) []any {
	values := make([]any, 0, length)
	for len(values) < length && this.err == nil {
		value, end := this.record()
		if end {
			this.fail()
			break
		}
		if nulls, ok := value.(nrbfNulls); ok {
			for range min(int(nulls), length-len(values)) {
				values = append(values, nil)
			}
			continue
		}
		values = append(values, value)
	}
	return values
}

func (this *nrbfStream) primitiveArray(primitive byte, length int) any {
	if primitive == primitiveByte {
		return append([]byte(nil), this.take(length)...)
	}
	values := make([]any, length)
	for i := range values {
		values[i] = this.primitive(primitive)
		if this.err != nil {
			break
		}
	}
	return values
}

func (this *nrbfStream) primitive(primitive byte) any {
	switch primitive {
	case primitiveBoolean:
		return this.u8() != 0
	case primitiveByte:
		return this.u8()
	case primitiveSByte:
		return int8(this.u8())
	case primitiveChar:
		if this.pos >= len(this.data) {
			this.fail()
			return nil
		}
		first := this.data[this.pos]
		size := 1
		switch {
		case first >= 0xF0:
			size = 4
		case first >= 0xE0:
			size = 3
		case first >= 0xC0:
			size = 2
		}
		r, _ := utf8.DecodeRune(this.take(size))
		return r
	case primitiveInt16:
		return int16(this.u16())
	case primitiveUInt16:
		return this.u16()
	case primitiveInt32:
		return this.i32()
	case primitiveUInt32:
		return this.u32()
	case primitiveInt64, primitiveTimeSpan, primitiveDateTime:
		return int64(this.u64())
	case primitiveUInt64:
		return this.u64()
	case primitiveSingle:
		return math.Float32frombits(this.u32())
	case primitiveDouble:
		return math.Float64frombits(this.u64())
	case primitiveDecimal, primitiveString:
		return this.text()
	case primitiveNull:
		return nil
	}
	this.fail()
	return nil
}

// value resolves a member reference.
func (this *nrbfStream) value(value any) any {
	for depth := 0; depth < 8; depth++ {
		reference, ok := value.(nrbfReference)
		if !ok {
			return value
		}
		value = this.objects[int32(reference)]
	}
	return nil
}
