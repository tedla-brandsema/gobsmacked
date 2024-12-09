package gobsmacked

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash/crc32"
	"math"
	"sync"
	"time"
)

const (
	BinaryFormatIdentifierString = "gobs"
	FileExtension                = "." + BinaryFormatIdentifierString
)

// File Prefix
const (
	formatFileMetaBytes     = len(BinaryFormatIdentifierString)
	versionFileMetaBytes    = 4
	checksumFileMetaBytes   = 4
	encryptedFileMetaBytes  = 1
	compressedFileMetaBytes = 1

	filePrefixSize = formatFileMetaBytes +
		versionFileMetaBytes +
		checksumFileMetaBytes +
		encryptedFileMetaBytes +
		compressedFileMetaBytes
)

var fileMeta = make([]byte, filePrefixSize)

// GOB Prefix
const (
	sizeMetaBytes      = 4 // fits math.MaxInt32
	checksumMetaBytes  = 4
	timestampMetaBytes = 8

	sizeStart = 0
	sizeEnd   = sizeMetaBytes

	checksumStart = sizeEnd
	checksumEnd   = checksumStart + checksumMetaBytes

	timestampStart = checksumEnd
	timestampEnd   = checksumEnd + timestampMetaBytes

	gobPrefixSize = sizeMetaBytes +
		checksumMetaBytes +
		timestampMetaBytes
)

const (
	// NOTE: because len() and count() return int, which size depends on the underlying architecture,
	// we set the maxTotalBytes to the lowest common denominator: math.MaxInt32,
	// of which we reserve gobPrefixSize, resulting in maxDataBytes.
	// In conclusion: the size of  a GOB is limited to maxDataBytes.
	maxTotalBytes = math.MaxInt32
	maxDataBytes  = maxTotalBytes - gobPrefixSize
)

func encodeGob(obj interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(obj); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decodeGob(data []byte, obj interface{}) error {
	buffer := bytes.NewReader(data)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(obj)
}

func gobPrefix(data []byte) ([gobPrefixSize]byte, error) {
	var err error

	var prefix [gobPrefixSize]byte

	var size [sizeMetaBytes]byte
	var checksum [checksumMetaBytes]byte
	var timestamp [timestampMetaBytes]byte

	var wg sync.WaitGroup
	wg.Add(3)

	// size prefix
	go func() {
		defer wg.Done()
		localSize, localErr := sizeMeta(data)
		if localErr != nil {
			err = localErr
			return
		}
		size = localSize
	}()

	// checksum prefix
	go func() {
		defer wg.Done()
		checksum = checksumMeta(data)
	}()

	// timestamp prefix
	go func() {
		defer wg.Done()
		timestamp = timestampMeta()
	}()

	wg.Wait()

	if err != nil {
		return prefix, err
	}

	// Assemble the prefix
	copy(prefix[sizeStart:sizeEnd], size[:])
	copy(prefix[checksumStart:checksumEnd], checksum[:])
	copy(prefix[timestampStart:timestampEnd], timestamp[:])

	return prefix, nil
}

func sizeMeta(data []byte) ([sizeMetaBytes]byte, error) {
	meta := [sizeMetaBytes]byte{}

	dataSize := len(data)
	if dataSize > maxDataBytes {
		return meta, fmt.Errorf("maximum allowed bytes %d exceeded: found %d", maxDataBytes, dataSize)
	}
	binary.LittleEndian.PutUint32(meta[:], uint32(dataSize))

	return meta, nil
}

func checksumMeta(data []byte) [checksumMetaBytes]byte {
	meta := [checksumMetaBytes]byte{}

	sum := crc32.ChecksumIEEE(data)
	binary.LittleEndian.PutUint32(meta[:], sum)

	return meta
}

func timestampMeta() [timestampMetaBytes]byte {
	meta := [timestampMetaBytes]byte{}

	now := time.Now().Unix()
	binary.LittleEndian.PutUint64(meta[:], uint64(now)) // time.Now() cannot be negative, safe conversion

	return meta
}
