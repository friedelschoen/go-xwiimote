package miidata

import (
	"errors"
	"io"
	"slices"
)

const (
	blockOffset = 0x0FCA
	blockSize   = 0x02f0
)

var MiiBlockMagic = []byte("RNCD")

var (
	ErrMagicMismatch = errors.New("block does not contain magic")
	ErrCRCMismatch   = errors.New("block does not match CRC")
)

func ReadMiiBlock(r io.ReaderAt) (dest []byte, err error) {
	var buf [752]byte
	for i := range 2 {
		err = getMiiBlock(r, buf[:], int64(i))
		if err == nil {
			dest = buf[:]
			return
		}
	}
	return
}

/**
 * Calculate a modified CRC16-CCITT checksum of a byte array, as used for
 * checking the validity of a Mii data block stored on a Wiimote.
 *
 * @param bytes the byte array to calculate the checksum for
 * @return the checksum (in the lower 16 bits)
 */
func ccittCRC16(bytes []byte) uint16 {
	crc := 0x0000
	for byteIndex := range bytes {
		for bitIndex := 7; bitIndex >= 0; bitIndex-- {
			left := (crc << 1) | ((int(bytes[byteIndex]) >> bitIndex) & 0x1)
			right := 0
			if crc&0x8000 != 0 {
				right = 0x1021
			}
			crc = left ^ right
		}
	}
	for counter := 16; counter > 0; counter-- {
		left := crc << 1
		right := 0
		if crc&0x8000 != 0 {
			right = 0x1021
		}

		crc = left ^ right
	}
	return uint16(crc & 0xFFFF)
}

func getMiiBlock(r io.ReaderAt, dest []byte, index int64) error {
	n, err := r.ReadAt(dest[:752], blockOffset+index*blockSize)
	if err != nil {
		return err
	}
	if n != 752 {
		return io.ErrUnexpectedEOF
	}
	if !slices.Equal(dest[0:4], MiiBlockMagic) {
		return ErrMagicMismatch
	}

	crc := ccittCRC16(dest[:750])
	if dest[750] != byte(crc>>8) || dest[751] != byte(crc) {
		return ErrCRCMismatch
	}
	return nil
}
