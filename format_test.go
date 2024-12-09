package gobsmacked

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"
	"time"
)

func TestSizeMeta(t *testing.T) {
	data := make([]byte, 100) // 100 bytes of data

	meta, err := sizeMeta(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := make([]byte, 4)
	binary.LittleEndian.PutUint32(expected, 100)

	if !bytes.Equal(meta[:], expected) {
		t.Errorf("Size meta mismatch. Got %v, expected %v", meta, expected)
	}
}

func TestSizeMeta_ExceedsMax(t *testing.T) {
	data := make([]byte, maxDataBytes+1) // Exceeds max data bytes

	_, err := sizeMeta(data)
	if err == nil {
		t.Fatal("Expected error, but got none")
	}
}

func TestChecksumMeta(t *testing.T) {
	data := []byte("test data")
	meta := checksumMeta(data)

	expected := make([]byte, 4)
	sum := crc32.ChecksumIEEE(data)
	binary.LittleEndian.PutUint32(expected, sum)

	if !bytes.Equal(meta[:], expected) {
		t.Errorf("Checksum meta mismatch. Got %v, expected %v", meta, expected)
	}
}

func TestTimestampMeta(t *testing.T) {
	before := uint64(time.Now().Unix())
	meta := timestampMeta()
	after := uint64(time.Now().Unix())

	timestamp := binary.LittleEndian.Uint64(meta[:])

	if timestamp < before || timestamp > after {
		t.Errorf("Timestamp out of expected range. Got %v, expected between %v and %v", timestamp, before, after)
	}
}

func TestGobMeta(t *testing.T) {
	data := []byte("test data")

	meta, err := gobPrefix(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	sizeMeta, err := sizeMeta(data)
	if err != nil {
		t.Fatalf("Unexpected error calculating size meta: %v", err)
	}
	checksumMeta := checksumMeta(data)
	timestampMeta := timestampMeta()

	if !bytes.Equal(meta[sizeStart:sizeEnd], sizeMeta[:]) {
		t.Errorf("Size meta mismatch in GobPrefix. Got %v, expected %v", meta[sizeStart:sizeEnd], sizeMeta[:])
	}
	if !bytes.Equal(meta[checksumStart:checksumEnd], checksumMeta[:]) {
		t.Errorf("Checksum meta mismatch in GobPrefix. Got %v, expected %v", meta[checksumStart:checksumEnd], checksumMeta[:])
	}
	if !bytes.Equal(meta[timestampStart:timestampEnd], timestampMeta[:]) {
		t.Errorf("Timestamp meta mismatch in GobPrefix. Got %v, expected %v", meta[timestampStart:timestampEnd], timestampMeta[:])
	}
}
