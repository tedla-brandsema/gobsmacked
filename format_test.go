package gobsmacked

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"
	"time"
)

func TestSizePrefix(t *testing.T) {
	data := make([]byte, 100) // 100 bytes of data

	prefix, err := sizePrefix(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := make([]byte, 4)
	binary.LittleEndian.PutUint32(expected, 100)

	if !bytes.Equal(prefix[:], expected) {
		t.Errorf("Size prefix mismatch. Got %v, expected %v", prefix, expected)
	}
}

func TestSizePrefix_ExceedsMax(t *testing.T) {
	data := make([]byte, maxDataBytes+1) // Exceeds max data bytes

	_, err := sizePrefix(data)
	if err == nil {
		t.Fatal("Expected error, but got none")
	}
}

func TestChecksumPrefix(t *testing.T) {
	data := []byte("test data")
	prefix := checksumPrefix(data)

	expected := make([]byte, 4)
	sum := crc32.ChecksumIEEE(data)
	binary.LittleEndian.PutUint32(expected, sum)

	if !bytes.Equal(prefix[:], expected) {
		t.Errorf("Checksum prefix mismatch. Got %v, expected %v", prefix, expected)
	}
}

func TestTimestampPrefix(t *testing.T) {
	before := uint64(time.Now().Unix())
	prefix := timestampPrefix()
	after := uint64(time.Now().Unix())

	timestamp := binary.LittleEndian.Uint64(prefix[:])

	if timestamp < before || timestamp > after {
		t.Errorf("Timestamp out of expected range. Got %v, expected between %v and %v", timestamp, before, after)
	}
}

func TestGobPrefix(t *testing.T) {
	data := []byte("test data")

	prefix, err := gobPrefix(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	sizePrefix, err := sizePrefix(data)
	if err != nil {
		t.Fatalf("Unexpected error calculating size prefix: %v", err)
	}
	checksumPrefix := checksumPrefix(data)
	timestampPrefix := timestampPrefix()

	if !bytes.Equal(prefix[sizeStart:sizeEnd], sizePrefix[:]) {
		t.Errorf("Size prefix mismatch in GobPrefix. Got %v, expected %v", prefix[sizeStart:sizeEnd], sizePrefix[:])
	}
	if !bytes.Equal(prefix[checksumStart:checksumEnd], checksumPrefix[:]) {
		t.Errorf("Checksum prefix mismatch in GobPrefix. Got %v, expected %v", prefix[checksumStart:checksumEnd], checksumPrefix[:])
	}
	if !bytes.Equal(prefix[timestampStart:timestampEnd], timestampPrefix[:]) {
		t.Errorf("Timestamp prefix mismatch in GobPrefix. Got %v, expected %v", prefix[timestampStart:timestampEnd], timestampPrefix[:])
	}
}
