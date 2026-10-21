package show

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
)

// Show data location in the cube's flash, as written by the vendor software into Update.hex.
const (
	HexShowStart = 0x0800B000
	HexShowEnd   = HexShowStart + blob.MaxSize // exclusive
)

var (
	ErrHexFormat   = errors.New("invalid Intel HEX record")
	ErrHexChecksum = errors.New("Intel HEX checksum mismatch in the show data")
	ErrHexNoShow   = errors.New("no show data in the Intel HEX file")
)

// ShowDataFromHex extracts the show data from an Intel HEX firmware image: the contiguous bytes
// from HexShowStart. Records outside the show data are not checked.
func ShowDataFromHex(data []byte) ([]byte, error) {
	var (
		buffer  = make([]byte, blob.MaxSize)
		present = make([]bool, blob.MaxSize)
		upper   uint32
	)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line[0] != ':' || len(line) < 11 || len(line)%2 == 0 {
			return nil, ErrHexFormat
		}
		record, err := hex.DecodeString(line[1:])
		if err != nil || len(record) != 5+int(record[0]) {
			return nil, ErrHexFormat
		}

		length := int(record[0])
		address := upper + uint32(record[1])<<8 + uint32(record[2])
		payload := record[4 : 4+length]
		switch record[3] {
		case 0x00:
			end := address + uint32(length)
			if end <= HexShowStart || address >= HexShowEnd {
				continue
			}
			if !validChecksum(record) {
				return nil, ErrHexChecksum
			}
			for i, value := range payload {
				if offset := address + uint32(i) - HexShowStart; offset < blob.MaxSize {
					buffer[offset], present[offset] = value, true
				}
			}
		case 0x01:
			return showRun(buffer, present)
		case 0x04:
			if length != 2 {
				return nil, ErrHexFormat
			}
			upper = uint32(payload[0])<<24 | uint32(payload[1])<<16
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return showRun(buffer, present)
}

func validChecksum(record []byte) bool {
	var sum byte
	for _, value := range record {
		sum += value
	}
	// The vendor software computes the checksum of its last, shorter record with length 16.
	return sum == 0 || sum+16-record[0] == 0
}

func showRun(buffer []byte, present []bool) ([]byte, error) {
	length := 0
	for length < len(present) && present[length] {
		length++
	}
	if length == 0 {
		return nil, ErrHexNoShow
	}
	return buffer[:length], nil
}
