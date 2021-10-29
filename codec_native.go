package avro

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"time"
	"unsafe"

	"github.com/modern-go/reflect2"
)

func createDecoderOfNative(schema Schema, typ reflect2.Type) ValDecoder {
	switch typ.Kind() {
	case reflect.Bool:
		if schema.Type() != Boolean {
			break
		}
		return &boolCodec{}

	case reflect.Int:
		if schema.Type() != Int {
			break
		}
		return &intCodec{}

	case reflect.Int8:
		if schema.Type() != Int {
			break
		}
		return &int8Codec{}

	case reflect.Int16:
		if schema.Type() != Int {
			break
		}
		return &int16Codec{}

	case reflect.Int32:
		if schema.Type() != Int {
			break
		}
		return &int32Codec{}

	case reflect.Int64:
		st := schema.Type()
		lt := getLogicalType(schema)
		switch {
		case st == Int && lt == TimeMillis: // time.Duration
			return &timeMillisCodec{}

		case st == Long && lt == TimeMicros: // time.Duration
			return &timeMicrosCodec{}

		case st == Long:
			return &int64Codec{}

		default:
			break
		}

	case reflect.Float32:
		if schema.Type() != Float {
			break
		}
		return &float32Codec{}

	case reflect.Float64:
		switch schema.Type() {
		// XL4:  Treat a Float64 as double or as int
		case Double:
			return &float64Codec{}

		case Int:
			return &int64Codec{}
		}

	case reflect.String:
		if schema.Type() != String {
			break
		}
		return &stringCodec{}

	case reflect.Slice:
		if typ.(reflect2.SliceType).Elem().Kind() != reflect.Uint8 || schema.Type() != Bytes {
			break
		}
		return &bytesCodec{sliceType: typ.(*reflect2.UnsafeSliceType)}

	case reflect.Struct:
		st := schema.Type()
		ls := getLogicalSchema(schema)
		lt := getLogicalType(schema)
		switch {
		case typ.RType() == timeRType && st == Int && lt == Date:
			return &dateCodec{}

		case typ.RType() == timeRType && st == Long && lt == TimestampMillis:
			return &timestampMillisCodec{}

		case typ.RType() == timeRType && st == Long && lt == TimestampMicros:
			return &timestampMicrosCodec{}

		case typ.RType() == ratRType && st == Bytes && lt == Decimal:
			dec := ls.(*DecimalLogicalSchema)

			return &bytesDecimalCodec{prec: dec.Precision(), scale: dec.Scale()}

		default:
			break
		}
	case reflect.Ptr:
		ptrType := typ.(*reflect2.UnsafePtrType)
		elemType := ptrType.Elem()

		ls := getLogicalSchema(schema)
		if ls == nil {
			break
		}
		if elemType.RType() != ratRType || schema.Type() != Bytes || ls.Type() != Decimal {
			break
		}
		dec := ls.(*DecimalLogicalSchema)

		return &bytesDecimalPtrCodec{prec: dec.Precision(), scale: dec.Scale()}
	}

	return &errorDecoder{err: fmt.Errorf("avro: %s is unsupported for Avro %s", typ.String(), schema.Type())}
}

func createEncoderOfNative(schema Schema, typ reflect2.Type) ValEncoder {
	switch typ.Kind() {
	case reflect.Bool:
		if schema.Type() != Boolean {
			break
		}
		return &boolCodec{}

	case reflect.Int:
		if schema.Type() != Int {
			break
		}
		return &intCodec{}

	case reflect.Int8:
		if schema.Type() != Int {
			break
		}
		return &int8Codec{}

	case reflect.Int16:
		if schema.Type() != Int {
			break
		}
		return &int16Codec{}

	case reflect.Int32:
		switch schema.Type() {
		case Long:
			return &int32LongCodec{}

		case Int:
			return &int32Codec{}
		}

	case reflect.Int64:
		st := schema.Type()
		lt := getLogicalType(schema)
		switch {
		case st == Int && lt == TimeMillis: // time.Duration
			return &timeMillisCodec{}

		case st == Long && lt == TimeMicros: // time.Duration
			return &timeMicrosCodec{}

		case st == Long:
			return &int64Codec{}

		default:
			break
		}

	case reflect.Float32:
		switch schema.Type() {
		case Double:
			return &float32DoubleCodec{}

		case Float:
			return &float32Codec{}
		}

	case reflect.Float64:
		switch schema.Type() {
		// XL4:  Treat a Float64 as double or long or int
		case Double:
			return &float64Codec{}

		case Long:
			return &longFromFloat64Codec{}

		case Int:
			return &intFromFloat64Codec{}
		}

	case reflect.String:
		if schema.Type() != String {
			break
		}
		return &stringCodec{}

	case reflect.Slice:
		if typ.(reflect2.SliceType).Elem().Kind() != reflect.Uint8 || schema.Type() != Bytes {
			break
		}
		return &bytesCodec{sliceType: typ.(*reflect2.UnsafeSliceType)}

	case reflect.Struct:
		st := schema.Type()
		lt := getLogicalType(schema)
		switch {
		case typ.RType() == timeRType && st == Int && lt == Date:
			return &dateCodec{}

		case typ.RType() == timeRType && st == Long && lt == TimestampMillis:
			return &timestampMillisCodec{}

		case typ.RType() == timeRType && st == Long && lt == TimestampMicros:
			return &timestampMicrosCodec{}

		case typ.RType() == ratRType && st != Bytes || lt == Decimal:
			ls := getLogicalSchema(schema)
			dec := ls.(*DecimalLogicalSchema)

			return &bytesDecimalCodec{prec: dec.Precision(), scale: dec.Scale()}

		default:
			break
		}

	case reflect.Ptr:
		ptrType := typ.(*reflect2.UnsafePtrType)
		elemType := ptrType.Elem()

		ls := getLogicalSchema(schema)
		if ls == nil {
			break
		}
		if elemType.RType() != ratRType || schema.Type() != Bytes || ls.Type() != Decimal {
			break
		}
		dec := ls.(*DecimalLogicalSchema)

		return &bytesDecimalPtrCodec{prec: dec.Precision(), scale: dec.Scale()}
	}

	if schema.Type() == Null {
		return &nullCodec{}
	}

	return &errorEncoder{err: fmt.Errorf("avro: %s is unsupported for Avro %s", typ.String(), schema.Type())}
}

func getLogicalSchema(schema Schema) LogicalSchema {
	lts, ok := schema.(LogicalTypeSchema)
	if !ok {
		return nil
	}

	return lts.Logical()
}

func getLogicalType(schema Schema) LogicalType {
	ls := getLogicalSchema(schema)
	if ls == nil {
		return ""
	}

	return ls.Type()
}

type nullCodec struct{}

func (*nullCodec) Encode(ptr unsafe.Pointer, w *Writer) {}

type boolCodec struct{}

func (*boolCodec) Decode(ptr unsafe.Pointer, r *Reader) {
	*((*bool)(ptr)) = r.ReadBool()
}

func (*boolCodec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteBool(*((*bool)(ptr)))
}

type intCodec struct{}

func (*intCodec) Decode(ptr unsafe.Pointer, r *Reader) {
	*((*int)(ptr)) = int(r.ReadInt())
}

func (*intCodec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteInt(int32(*((*int)(ptr))))
}

type int8Codec struct{}

func (*int8Codec) Decode(ptr unsafe.Pointer, r *Reader) {
	*((*int8)(ptr)) = int8(r.ReadInt())
}

func (*int8Codec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteInt(int32(*((*int8)(ptr))))
}

type int16Codec struct{}

func (*int16Codec) Decode(ptr unsafe.Pointer, r *Reader) {
	*((*int16)(ptr)) = int16(r.ReadInt())
}

func (*int16Codec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteInt(int32(*((*int16)(ptr))))
}

type int32Codec struct{}

func (*int32Codec) Decode(ptr unsafe.Pointer, r *Reader) {
	*((*int32)(ptr)) = r.ReadInt()
}

func (*int32Codec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteInt(*((*int32)(ptr)))
}

type int32LongCodec struct{}

func (*int32LongCodec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteLong(int64(*((*int32)(ptr))))
}

type int64Codec struct{}

func (*int64Codec) Decode(ptr unsafe.Pointer, r *Reader) {
	*((*int64)(ptr)) = r.ReadLong()
}

func (*int64Codec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteLong(*((*int64)(ptr)))
}

type float32Codec struct{}

func (*float32Codec) Decode(ptr unsafe.Pointer, r *Reader) {
	*((*float32)(ptr)) = r.ReadFloat()
}

func (*float32Codec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteFloat(*((*float32)(ptr)))
}

type float32DoubleCodec struct{}

func (*float32DoubleCodec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteDouble(float64(*((*float32)(ptr))))
}

type float64Codec struct{}

func (*float64Codec) Decode(ptr unsafe.Pointer, r *Reader) {
	*((*float64)(ptr)) = r.ReadDouble()
}

func (*float64Codec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteDouble(*((*float64)(ptr)))
}

type intFromFloat64Codec struct{}

func (*intFromFloat64Codec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteInt(int32(*((*float64)(ptr))))
}

type longFromFloat64Codec struct{}

func (*longFromFloat64Codec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteLong(int64(*((*float64)(ptr))))
}

type stringCodec struct{}

func (*stringCodec) Decode(ptr unsafe.Pointer, r *Reader) {
	*((*string)(ptr)) = r.ReadString()
}

func (*stringCodec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteString(*((*string)(ptr)))
}

type bytesCodec struct {
	sliceType *reflect2.UnsafeSliceType
}

func (c *bytesCodec) Decode(ptr unsafe.Pointer, r *Reader) {
	b := r.ReadBytes()
	c.sliceType.UnsafeSet(ptr, reflect2.PtrOf(b))
}

func (c *bytesCodec) Encode(ptr unsafe.Pointer, w *Writer) {
	w.WriteBytes(*((*[]byte)(ptr)))
}

type dateCodec struct{}

func (c *dateCodec) Decode(ptr unsafe.Pointer, r *Reader) {
	i := r.ReadInt()
	*((*time.Time)(ptr)) = time.Unix(0, int64(i)*int64(24*time.Hour)).UTC()
}

func (c *dateCodec) Encode(ptr unsafe.Pointer, w *Writer) {
	t := *((*time.Time)(ptr))
	w.WriteInt(int32(t.UnixNano() / int64(24*time.Hour)))
}

type timestampMillisCodec struct{}

func (c *timestampMillisCodec) Decode(ptr unsafe.Pointer, r *Reader) {
	i := r.ReadLong()
	sec := i / 1e3
	nsec := (i - sec*1e3) * 1e6
	*((*time.Time)(ptr)) = time.Unix(sec, nsec).UTC()
}

func (c *timestampMillisCodec) Encode(ptr unsafe.Pointer, w *Writer) {
	t := *((*time.Time)(ptr))
	w.WriteLong(t.Unix()*1e3 + int64(t.Nanosecond()/1e6))
}

type timestampMicrosCodec struct{}

func (c *timestampMicrosCodec) Decode(ptr unsafe.Pointer, r *Reader) {
	i := r.ReadLong()
	sec := i / 1e6
	nsec := (i - sec*1e6) * 1e3
	*((*time.Time)(ptr)) = time.Unix(sec, nsec).UTC()
}

func (c *timestampMicrosCodec) Encode(ptr unsafe.Pointer, w *Writer) {
	t := *((*time.Time)(ptr))
	w.WriteLong(t.Unix()*1e6 + int64(t.Nanosecond()/1e3))
}

type timeMillisCodec struct{}

func (c *timeMillisCodec) Decode(ptr unsafe.Pointer, r *Reader) {
	i := r.ReadInt()
	*((*time.Duration)(ptr)) = time.Duration(i) * time.Millisecond
}

func (c *timeMillisCodec) Encode(ptr unsafe.Pointer, w *Writer) {
	d := *((*time.Duration)(ptr))
	w.WriteInt(int32(d.Nanoseconds() / int64(time.Millisecond)))
}

type timeMicrosCodec struct{}

func (c *timeMicrosCodec) Decode(ptr unsafe.Pointer, r *Reader) {
	i := r.ReadLong()
	*((*time.Duration)(ptr)) = time.Duration(i) * time.Microsecond
}

func (c *timeMicrosCodec) Encode(ptr unsafe.Pointer, w *Writer) {
	d := *((*time.Duration)(ptr))
	w.WriteLong(d.Nanoseconds() / int64(time.Microsecond))
}

var one = big.NewInt(1)

type bytesDecimalCodec struct {
	prec  int
	scale int
}

func (c *bytesDecimalCodec) Decode(ptr unsafe.Pointer, r *Reader) {
	b := r.ReadBytes()
	if i := (&big.Int{}).SetBytes(b); len(b) > 0 && b[0]&0x80 > 0 {
		i.Sub(i, new(big.Int).Lsh(one, uint(len(b))*8))
	}
	*((*big.Rat)(ptr)) = *ratFromBytes(b, c.scale)
}

func ratFromBytes(b []byte, scale int) *big.Rat {
	i := (&big.Int{}).SetBytes(b)
	if len(b) > 0 && b[0]&0x80 > 0 {
		i.Sub(i, new(big.Int).Lsh(one, uint(len(b))*8))
	}
	return big.NewRat(i.Int64(), int64(math.Pow10(scale)))
}

func (c *bytesDecimalCodec) Encode(ptr unsafe.Pointer, w *Writer) {
	r := (*big.Rat)(ptr)
	i := (&big.Int{}).Mul(r.Num(), big.NewInt(int64(math.Pow10(c.scale))))
	i = i.Div(i, r.Denom())

	var b []byte
	switch i.Sign() {
	case 0:
		b = []byte{0}

	case 1:
		b = i.Bytes()
		if b[0]&0x80 > 0 {
			b = append([]byte{0}, b...)
		}

	case -1:
		length := uint(i.BitLen()/8+1) * 8
		b = i.Add(i, (&big.Int{}).Lsh(one, length)).Bytes()
	}
	w.WriteBytes(b)
}

type bytesDecimalPtrCodec struct {
	prec  int
	scale int
}

func (c *bytesDecimalPtrCodec) Decode(ptr unsafe.Pointer, r *Reader) {
	b := r.ReadBytes()
	if i := (&big.Int{}).SetBytes(b); len(b) > 0 && b[0]&0x80 > 0 {
		i.Sub(i, new(big.Int).Lsh(one, uint(len(b))*8))
	}
	*((**big.Rat)(ptr)) = ratFromBytes(b, c.scale)
}

func (c *bytesDecimalPtrCodec) Encode(ptr unsafe.Pointer, w *Writer) {
	r := *((**big.Rat)(ptr))
	i := (&big.Int{}).Mul(r.Num(), big.NewInt(int64(math.Pow10(c.scale))))
	i = i.Div(i, r.Denom())

	var b []byte
	switch i.Sign() {
	case 0:
		b = []byte{0}

	case 1:
		b = i.Bytes()
		if b[0]&0x80 > 0 {
			b = append([]byte{0}, b...)
		}

	case -1:
		length := uint(i.BitLen()/8+1) * 8
		b = i.Add(i, (&big.Int{}).Lsh(one, length)).Bytes()
	}
	w.WriteBytes(b)
}
