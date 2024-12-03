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

// File Meta
const (
	formatFilePrefixBytes     = len(BinaryFormatIdentifierString)
	versionFilePrefixBytes    = 4
	checksumFilePrefixBytes   = 4
	encryptedFilePrefixBytes  = 1
	compressedFilePrefixBytes = 1

	fileMetaSize = formatFilePrefixBytes +
		versionFilePrefixBytes +
		checksumFilePrefixBytes +
		encryptedFilePrefixBytes +
		compressedFilePrefixBytes
)

var fileMeta = make([]byte, fileMetaSize)

// GOB Meta
const (
	sizePrefixBytes      = 4 // fits math.MaxInt32
	checksumPrefixBytes  = 4
	timestampPrefixBytes = 8

	sizeStart = 0
	sizeEnd   = sizePrefixBytes

	checksumStart = sizeEnd
	checksumEnd   = checksumStart + checksumPrefixBytes

	timestampStart = checksumEnd
	timestampEnd   = checksumEnd + timestampPrefixBytes

	gobMetaSize = sizePrefixBytes +
		checksumPrefixBytes +
		timestampPrefixBytes
)

const (
	// NOTE: because len() and count() return int, which size depends on the underlying architecture,
	// we set the maxTotalBytes to the lowest common denominator: math.MaxInt32,
	// of which we reserve gobMetaSize, resulting in maxDataBytes.
	// In conclusion: the size of  a GOB is limited to maxDataBytes.
	maxTotalBytes = math.MaxInt32
	maxDataBytes  = maxTotalBytes - gobMetaSize
)

func encodeGob(obj interface{}) ([]byte, error) {
	// TODO: do we append the GOBS file with a map of GOB types?
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

func gobPrefix(data []byte) ([gobMetaSize]byte, error) {
	var err error

	var prefix [gobMetaSize]byte

	var size [sizePrefixBytes]byte
	var checksum [checksumPrefixBytes]byte
	var timestamp [timestampPrefixBytes]byte

	var wg sync.WaitGroup
	wg.Add(3)

	// size prefix
	go func() {
		defer wg.Done()
		localSize, localErr := sizePrefix(data)
		if localErr != nil {
			err = localErr
			return
		}
		size = localSize
	}()

	// checksum prefix
	go func() {
		defer wg.Done()
		checksum = checksumPrefix(data)
	}()

	// timestamp prefix
	go func() {
		defer wg.Done()
		timestamp = timestampPrefix()
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

func sizePrefix(data []byte) ([sizePrefixBytes]byte, error) {
	prefix := [sizePrefixBytes]byte{}

	dataSize := len(data)
	if dataSize > maxDataBytes {
		return prefix, fmt.Errorf("maximum allowed bytes %d exceeded: found %d", maxDataBytes, dataSize)
	}
	binary.LittleEndian.PutUint32(prefix[:], uint32(dataSize))

	return prefix, nil
}

func checksumPrefix(data []byte) [checksumPrefixBytes]byte {
	prefix := [checksumPrefixBytes]byte{}

	sum := crc32.ChecksumIEEE(data)
	binary.LittleEndian.PutUint32(prefix[:], sum)

	return prefix
}

func timestampPrefix() [timestampPrefixBytes]byte {
	prefix := [timestampPrefixBytes]byte{}

	now := time.Now().Unix()
	binary.LittleEndian.PutUint64(prefix[:], uint64(now)) // time.Now() cannot be negative, safe conversion

	return prefix
}
