package avro_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/xl4hub/hamba-avro"
	"github.com/stretchr/testify/assert"
)

func TestDecoder_FixedInvalidType(t *testing.T) {
	defer ConfigTeardown()

	data := []byte{0x66, 0x6F, 0x6F, 0x66, 0x6F, 0x6F}
	schema := `{"type":"fixed", "name": "test", "size": 6}`
	dec, err := avro.NewDecoder(schema, bytes.NewReader(data))
	assert.NoError(t, err)

	var i [6]int
	err = dec.Decode(&i)

	assert.Error(t, err)
}

func TestDecoder_Fixed(t *testing.T) {
	defer ConfigTeardown()

	data := []byte{0x66, 0x6F, 0x6F, 0x66, 0x6F, 0x6F}
	schema := `{"type":"fixed", "name": "test", "size": 6}`
	dec, _ := avro.NewDecoder(schema, bytes.NewReader(data))

	var got [6]byte
	err := dec.Decode(&got)

	assert.NoError(t, err)
	assert.Equal(t, [6]byte{'f', 'o', 'o', 'f', 'o', 'o'}, got)
}

func TestDecoder_FixedRat_Positive(t *testing.T) {
	defer ConfigTeardown()

	data := []byte{0x00, 0x00, 0x00, 0x00, 0x87, 0x78}
	schema := `{"type":"fixed", "name": "test", "size": 6,"logicalType":"decimal","precision":4,"scale":2}`
	dec, err := avro.NewDecoder(schema, bytes.NewReader(data))
	assert.NoError(t, err)

	got := &big.Rat{}
	err = dec.Decode(got)

	assert.NoError(t, err)
	assert.Equal(t, big.NewRat(1734, 5), got)
}

func TestDecoder_FixedRat_Negative(t *testing.T) {
	defer ConfigTeardown()

	data := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x78, 0x88}
	schema := `{"type":"fixed", "name": "test", "size": 6, "logicalType":"decimal","precision":4,"scale":2}`
	dec, err := avro.NewDecoder(schema, bytes.NewReader(data))
	assert.NoError(t, err)

	got := &big.Rat{}
	err = dec.Decode(got)

	assert.NoError(t, err)
	assert.Equal(t, big.NewRat(-1734, 5), got)
}

func TestDecoder_FixedRat_Zero(t *testing.T) {
	defer ConfigTeardown()

	data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	schema := `{"type":"fixed", "name": "test", "size": 6,"logicalType":"decimal","precision":4,"scale":2}`
	dec, err := avro.NewDecoder(schema, bytes.NewReader(data))
	assert.NoError(t, err)

	got := &big.Rat{}
	err = dec.Decode(got)

	assert.NoError(t, err)
	assert.Equal(t, big.NewRat(0, 1), got)
}

func TestDecoder_FixedRatInvalidLogicalSchema(t *testing.T) {
	defer ConfigTeardown()

	data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	schema := `{"type":"fixed", "name": "test", "size": 6}`
	dec, err := avro.NewDecoder(schema, bytes.NewReader(data))
	assert.NoError(t, err)

	got := &big.Rat{}
	err = dec.Decode(got)

	assert.Error(t, err)
}
