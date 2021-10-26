package ocf_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/xl4hub/hamba-avro/ocf"
	"github.com/stretchr/testify/assert"
)

var schema = `{
	"type":"record",
	"name":"FullRecord",
	"namespace":"org.hamba.avro",
	"fields":[
		{"name":"strings","type":{"type":"array","items":"string"}},
		{"name":"longs","type":{"type":"array","items":"long"}},
		{"name":"enum","type":{"type":"enum","name":"foo","symbols":["A","B","C","D"]}},
		{"name":"map","type":{"type":"map","values":"int"}},
		{"name":"nullable","type":["null","string"]},
		{"name":"fixed","type":{"type":"fixed","name":"md5","size":16}},
		{"name":"record","type":{
			"type":"record",
			"name":"TestRecord",
			"fields":[
				{"name":"long","type":"long"},
				{"name":"string","type":"string"},
				{"name":"int","type":"int"},
				{"name":"float","type":"float"},
				{"name":"double","type":"double"},
				{"name":"bool","type":"boolean"}
			]
		}}
	]
}`

type FullRecord struct {
	Strings  []string       `avro:"strings"`
	Longs    []int64        `avro:"longs"`
	Enum     string         `avro:"enum"`
	Map      map[string]int `avro:"map"`
	Nullable *string        `avro:"nullable"`
	Fixed    [16]byte       `avro:"fixed"`
	Record   *TestRecord    `avro:"record"`
}

type TestRecord struct {
	Long   int64   `avro:"long"`
	String string  `avro:"string"`
	Int    int32   `avro:"int"`
	Float  float32 `avro:"float"`
	Double float64 `avro:"double"`
	Bool   bool    `avro:"bool"`
}

func TestNewDecoder_InvalidHeader(t *testing.T) {
	data := []byte{'O', 'b', 'j'}

	_, err := ocf.NewDecoder(bytes.NewReader(data))

	assert.Error(t, err)
}

func TestNewDecoder_InvalidMagic(t *testing.T) {
	data := []byte{'f', 'o', 'o', 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	_, err := ocf.NewDecoder(bytes.NewReader(data))

	assert.Error(t, err)
}

func TestNewDecoder_InvalidSchema(t *testing.T) {
	data := []byte{'O', 'b', 'j', 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	_, err := ocf.NewDecoder(bytes.NewReader(data))

	assert.Error(t, err)
}

func TestNewDecoder_InvalidCodec(t *testing.T) {
	data := []byte{'O', 'b', 'j', 0x1, 0x3, 0x4c, 0x16, 'a', 'v', 'r', 'o', '.', 's', 'c', 'h', 'e', 'm', 'a', 0xc, 0x22, 'l', 'o', 'n', 'g',
		0x22, 0x14, 'a', 'v', 'r', 'o', 0x2e, 'c', 'o', 'd', 'e', 'c', 0xe, 'd', 'e', 'a', 'l', 'a', 't', 'e', 0x0,
		0x72, 0xce, 0x78, 0x7, 0x35, 0x81, 0xb0, 0x80, 0x77, 0x59, 0xa9, 0x83, 0xaf, 0x90, 0x3e, 0xaf,
	}

	_, err := ocf.NewDecoder(bytes.NewReader(data))

	assert.Error(t, err)
}

func TestDecoder(t *testing.T) {
	unionStr := "union value"
	want := FullRecord{
		Strings: []string{"string1", "string2", "string3", "string4", "string5"},
		Longs:   []int64{1, 2, 3, 4, 5},
		Enum:    "C",
		Map: map[string]int{
			"key1": 1,
			"key2": 2,
			"key3": 3,
			"key4": 4,
			"key5": 5,
		},
		Nullable: &unionStr,
		Fixed:    [16]byte{0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04},
		Record: &TestRecord{
			Long:   1925639126735,
			String: "I am a test record",
			Int:    666,
			Float:  7171.17,
			Double: 916734926348163.01973408746523,
			Bool:   true,
		},
	}

	f, err := os.Open("../testdata/full.avro")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	dec, err := ocf.NewDecoder(f)
	if err != nil {
		t.Error(err)
		return
	}

	var count int
	for dec.HasNext() {
		count++
		var got FullRecord
		err = dec.Decode(&got)

		assert.NoError(t, err)
		assert.Equal(t, want, got)
	}

	assert.NoError(t, dec.Error())
	assert.Equal(t, 1, count)
}

func TestDecoderDeflate(t *testing.T) {
	unionStr := "union value"
	want := FullRecord{
		Strings: []string{"string1", "string2", "string3", "string4", "string5"},
		Longs:   []int64{1, 2, 3, 4, 5},
		Enum:    "C",
		Map: map[string]int{
			"key1": 1,
			"key2": 2,
			"key3": 3,
			"key4": 4,
			"key5": 5,
		},
		Nullable: &unionStr,
		Fixed:    [16]byte{0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04},
		Record: &TestRecord{
			Long:   1925639126735,
			String: "I am a test record",
			Int:    666,
			Float:  7171.17,
			Double: 916734926348163.01973408746523,
			Bool:   true,
		},
	}

	f, err := os.Open("../testdata/full-deflate.avro")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	dec, err := ocf.NewDecoder(f)
	if err != nil {
		t.Error(err)
		return
	}

	var count int
	for dec.HasNext() {
		count++
		var got FullRecord
		err = dec.Decode(&got)

		assert.NoError(t, err)
		assert.Equal(t, want, got)
	}

	assert.NoError(t, dec.Error())
	assert.Equal(t, 1, count)
}

func TestDecoderDeflate_InvalidData(t *testing.T) {
	f, err := os.Open("../testdata/deflate-invalid-data.avro")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	dec, err := ocf.NewDecoder(f)
	if err != nil {
		t.Error(err)
		return
	}

	dec.HasNext()

	assert.Error(t, dec.Error())
}

func TestDecoderSnappy(t *testing.T) {
	unionStr := "union value"
	want := FullRecord{
		Strings: []string{"string1", "string2", "string3", "string4", "string5"},
		Longs:   []int64{1, 2, 3, 4, 5},
		Enum:    "C",
		Map: map[string]int{
			"key1": 1,
			"key2": 2,
			"key3": 3,
			"key4": 4,
			"key5": 5,
		},
		Nullable: &unionStr,
		Fixed:    [16]byte{0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04},
		Record: &TestRecord{
			Long:   1925639126735,
			String: "I am a test record",
			Int:    666,
			Float:  7171.17,
			Double: 916734926348163.01973408746523,
			Bool:   true,
		},
	}

	f, err := os.Open("../testdata/full-snappy.avro")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	dec, err := ocf.NewDecoder(f)
	if err != nil {
		t.Error(err)
		return
	}

	var count int
	for dec.HasNext() {
		count++
		var got FullRecord
		err = dec.Decode(&got)

		assert.NoError(t, err)
		assert.Equal(t, want, got)
	}

	assert.NoError(t, dec.Error())
	assert.Equal(t, 1, count)
}

func TestDecoderSnappy_InvalidData(t *testing.T) {
	f, err := os.Open("../testdata/snappy-invalid-data.avro")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	dec, err := ocf.NewDecoder(f)
	if err != nil {
		t.Error(err)
		return
	}

	dec.HasNext()

	assert.Error(t, dec.Error())
}

func TestDecoderSnappy_ShortCRC(t *testing.T) {
	f, err := os.Open("../testdata/snappy-short-crc.avro")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	dec, err := ocf.NewDecoder(f)
	if err != nil {
		t.Error(err)
		return
	}

	dec.HasNext()

	assert.Error(t, dec.Error())
}

func TestDecoderSnappy_InvalidCRC(t *testing.T) {
	f, err := os.Open("../testdata/snappy-invalid-crc.avro")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	dec, err := ocf.NewDecoder(f)
	if err != nil {
		t.Error(err)
		return
	}

	dec.HasNext()

	assert.Error(t, dec.Error())
}

func TestDecoder_DecodeAvroError(t *testing.T) {
	data := []byte{'O', 'b', 'j', 0x01, 0x01, 0x26, 0x16, 'a', 'v', 'r', 'o', '.', 's', 'c', 'h', 'e', 'm', 'a',
		0x0c, '"', 'l', 'o', 'n', 'g', '"', 0x00, 0xfb, 0x2b, 0x0f, 0x1a, 0xdd, 0xfd, 0x90, 0x7d, 0x87, 0x12,
		0x15, 0x29, 0xd7, 0x1d, 0x1c, 0xdd, 0x02, 0x16, 0xe2, 0xa2, 0xf3, 0xad, 0xad, 0xad, 0xe2, 0xa2, 0xf3,
		0xad, 0xad, 0xfb, 0x2b, 0x0f, 0x1a, 0xdd, 0xfd, 0x90, 0x7d, 0x87, 0x12, 0x15, 0x29, 0xd7, 0x1d, 0x1c, 0xdd,
	}

	dec, _ := ocf.NewDecoder(bytes.NewReader(data))
	_ = dec.HasNext()

	var l int64
	err := dec.Decode(&l)

	assert.Error(t, err)
}

func TestDecoder_DecodeMustCallHasNext(t *testing.T) {
	data := []byte{'O', 'b', 'j', 0x01, 0x01, 0x26, 0x16, 'a', 'v', 'r', 'o', '.', 's', 'c', 'h', 'e', 'm', 'a',
		0x0c, '"', 'l', 'o', 'n', 'g', '"', 0x00, 0xfb, 0x2b, 0x0f, 0x1a, 0xdd, 0xfd, 0x90, 0x7d, 0x87, 0x12,
		0x15, 0x29, 0xd7, 0x1d, 0x1c, 0xdd, 0x02, 0x02, 0x02, 0xfb, 0x2b, 0x0f, 0x1a, 0xdd, 0xfd, 0x90, 0x7d,
		0x87, 0x12, 0x15, 0x29, 0xd7, 0x1d, 0x1c, 0xdd,
	}

	dec, _ := ocf.NewDecoder(bytes.NewReader(data))

	var l int64
	err := dec.Decode(&l)

	assert.Error(t, err)
}

func TestDecoder_InvalidBlock(t *testing.T) {
	data := []byte{'O', 'b', 'j', 0x01, 0x01, 0x26, 0x16, 'a', 'v', 'r', 'o', '.', 's', 'c', 'h', 'e', 'm', 'a',
		0x0c, '"', 'l', 'o', 'n', 'g', '"', 0x00, 0xfa, 0x2b, 0x0f, 0x1a, 0xdd, 0xfd, 0x90, 0x7d, 0x87, 0x12,
		0x15, 0x29, 0xd7, 0x1d, 0x1c, 0xdd, 0x02, 0x02, 0x02, 0xfb, 0x2b, 0x0f, 0x1a, 0xdd, 0xfd, 0x90, 0x7d,
		0x87, 0x12, 0x15, 0x29, 0xd7, 0x1d, 0x1c, 0xdd,
	}

	dec, _ := ocf.NewDecoder(bytes.NewReader(data))

	got := dec.HasNext()

	assert.False(t, got)
	assert.Error(t, dec.Error())
}

func TestNewEncoder_InvalidSchema(t *testing.T) {
	buf := &bytes.Buffer{}

	_, err := ocf.NewEncoder(``, buf)

	assert.Error(t, err)
}

func TestNewEncoder_InvalidCodec(t *testing.T) {
	buf := &bytes.Buffer{}

	_, err := ocf.NewEncoder(`"long"`, buf, ocf.WithCodec(ocf.CodecName("test")))

	assert.Error(t, err)
}

func TestEncoder(t *testing.T) {
	unionStr := "union value"
	record := FullRecord{
		Strings: []string{"string1", "string2", "string3", "string4", "string5"},
		Longs:   []int64{1, 2, 3, 4, 5},
		Enum:    "C",
		Map: map[string]int{
			"key1": 1,
			"key2": 2,
			"key3": 3,
			"key4": 4,
			"key5": 5,
		},
		Nullable: &unionStr,
		Fixed:    [16]byte{0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04},
		Record: &TestRecord{
			Long:   1925639126735,
			String: "I am a test record",
			Int:    666,
			Float:  7171.17,
			Double: 916734926348163.01973408746523,
			Bool:   true,
		},
	}

	buf := &bytes.Buffer{}
	enc, err := ocf.NewEncoder(schema, buf)
	if err != nil {
		t.Error(err)
		return
	}

	err = enc.Encode(record)
	assert.NoError(t, err)

	err = enc.Close()
	assert.NoError(t, err)
}

func TestEncoder_EncodeCompressesDeflate(t *testing.T) {
	unionStr := "union value"
	record := FullRecord{
		Strings: []string{"string1", "string2", "string3", "string4", "string5"},
		Longs:   []int64{1, 2, 3, 4, 5},
		Enum:    "C",
		Map: map[string]int{
			"key1": 1,
			"key2": 2,
			"key3": 3,
			"key4": 4,
			"key5": 5,
		},
		Nullable: &unionStr,
		Fixed:    [16]byte{0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04},
		Record: &TestRecord{
			Long:   1925639126735,
			String: "I am a test record",
			Int:    666,
			Float:  7171.17,
			Double: 916734926348163.01973408746523,
			Bool:   true,
		},
	}

	buf := &bytes.Buffer{}
	enc, _ := ocf.NewEncoder(schema, buf, ocf.WithCodec(ocf.Deflate))
	defer enc.Close()

	err := enc.Encode(record)
	assert.NoError(t, err)

	err = enc.Close()
	assert.NoError(t, err)

	assert.Equal(t, 926, buf.Len())
}

func TestEncoder_EncodeCompressesSnappy(t *testing.T) {
	unionStr := "union value"
	record := FullRecord{
		Strings: []string{"string1", "string2", "string3", "string4", "string5"},
		Longs:   []int64{1, 2, 3, 4, 5},
		Enum:    "C",
		Map: map[string]int{
			"key1": 1,
			"key2": 2,
			"key3": 3,
			"key4": 4,
			"key5": 5,
		},
		Nullable: &unionStr,
		Fixed:    [16]byte{0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04},
		Record: &TestRecord{
			Long:   1925639126735,
			String: "I am a test record",
			Int:    666,
			Float:  7171.17,
			Double: 916734926348163.01973408746523,
			Bool:   true,
		},
	}

	buf := &bytes.Buffer{}
	enc, _ := ocf.NewEncoder(schema, buf, ocf.WithBlockLength(1), ocf.WithCodec(ocf.Snappy))

	err := enc.Encode(record)
	assert.NoError(t, err)

	err = enc.Close()
	assert.NoError(t, err)

	assert.Equal(t, 938, buf.Len())
}

func TestEncoder_EncodeError(t *testing.T) {
	buf := &bytes.Buffer{}
	enc, _ := ocf.NewEncoder(`"long"`, buf)

	err := enc.Encode("test")

	assert.Error(t, err)
}

func TestEncoder_EncodeWritesBlocks(t *testing.T) {
	buf := &bytes.Buffer{}
	enc, _ := ocf.NewEncoder(`"long"`, buf, ocf.WithBlockLength(1))
	defer enc.Close()

	err := enc.Encode(int64(1))

	assert.NoError(t, err)
	assert.Equal(t, 77, buf.Len())
}

func TestEncoder_EncodeHandlesWriteBlockError(t *testing.T) {
	w := &errorWriter{}
	enc, _ := ocf.NewEncoder(`"long"`, w, ocf.WithBlockLength(1))
	defer enc.Close()

	err := enc.Encode(int64(1))

	assert.Error(t, err)
}

func TestEncoder_CloseHandlesWriteBlockError(t *testing.T) {
	w := &errorWriter{}
	enc, _ := ocf.NewEncoder(`"long"`, w)
	_ = enc.Encode(int64(1))

	err := enc.Close()

	assert.Error(t, err)
}

func TestEncodeDecodeMetadata(t *testing.T) {
	buf := &bytes.Buffer{}
	enc, _ := ocf.NewEncoder(`"long"`, buf, ocf.WithMetadata(map[string][]byte{
		"test": []byte("foo"),
	}))

	err := enc.Encode(int64(1))
	assert.NoError(t, err)

	enc.Close()

	dec, err := ocf.NewDecoder(buf)
	assert.NoError(t, err)

	assert.Equal(t, []byte("foo"), dec.Metadata()["test"])
}

type errorWriter struct{}

func (*errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("test")
}
