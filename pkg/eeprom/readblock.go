package eeprom

import (
	"bytes"
	"errors"
	"io"

	"github.com/friedelschoen/go-wiimote"
)

const (
	MiiBlockOffset   = 0x0fca
	IRCalibOffset    = 0x0000
	AccelCalibOffset = 0x0016
	MiiBlockMagic    = "RNCD"
)

var (
	ErrMagicMismatch = errors.New("block does not contain magic")
	ErrCRCMismatch   = errors.New("block does not match CRC")
)

func readTwice(r io.ReaderAt, offset int64, dest []byte, magic []byte, chksum func([]byte) bool) error {
	var last error

	for i := range int64(2) {
		n, err := r.ReadAt(dest, offset+i*int64(len(dest)))

		if n != len(dest) {
			if n == 0 {
				continue
			}
			if err == nil {
				err = io.ErrUnexpectedEOF
			}
			last = err
			continue
		}

		if len(magic) > 0 && !bytes.HasPrefix(dest, magic) {
			last = ErrMagicMismatch
			continue
		}
		if chksum != nil && !chksum(dest) {
			last = ErrCRCMismatch
			continue
		}
		return nil
	}

	return last
}

func miiCRC(bytes []byte) bool {
	crc := uint16(0x0000)
	for _, chr := range bytes[:750] {
		for bitIndex := 7; bitIndex >= 0; bitIndex-- {
			left := (crc << 1) | ((uint16(chr) >> bitIndex) & 0x01)
			right := uint16(0)
			if crc&0x8000 != 0 {
				right = 0x1021
			}
			crc = left ^ right
		}
	}
	for counter := 16; counter > 0; counter-- {
		left := crc << 1
		right := uint16(0)
		if crc&0x8000 != 0 {
			right = 0x1021
		}

		crc = left ^ right
	}
	return bytes[750] == byte(crc>>8) && bytes[751] == byte(crc)
}

func ReadMiiBlock(r io.ReaderAt) (buf MiiBlock, err error) {
	err = readTwice(r, MiiBlockOffset, buf[:], []byte(MiiBlockMagic), miiCRC)
	return
}

func calibCRC(b []byte) bool {
	crc := uint8(0x55)
	for _, c := range b[:len(b)-1] {
		crc += c
	}
	return b[len(b)-1] == crc
}

func ReadIRCalibration(r io.ReaderAt) (slots [4]wiimote.Vec2, err error) {
	var buf [11]byte
	err = readTwice(r, IRCalibOffset, buf[:], nil, calibCRC)
	if err != nil {
		return
	}

	slots[0].X = int32(buf[0]) | (int32((buf[2]>>4)&0x03) << 8)
	slots[0].Y = int32(buf[1]) | (int32((buf[2]>>6)&0x03) << 8)
	slots[1].X = int32(buf[3]) | (int32((buf[2]>>0)&0x03) << 8)
	slots[1].Y = int32(buf[4]) | (int32((buf[2]>>2)&0x03) << 8)

	slots[2].X = int32(buf[5]) | (int32((buf[7]>>4)&0x03) << 8)
	slots[2].Y = int32(buf[6]) | (int32((buf[7]>>6)&0x03) << 8)
	slots[3].X = int32(buf[8]) | (int32((buf[7]>>0)&0x03) << 8)
	slots[3].Y = int32(buf[9]) | (int32((buf[7]>>2)&0x03) << 8)
	return
}

func ReadAccelCalibration(r io.ReaderAt) (accel [2]wiimote.Vec3, speaker uint8, motor bool, err error) {
	var buf [10]byte
	err = readTwice(r, AccelCalibOffset, buf[:], nil, calibCRC)
	if err != nil {
		return
	}

	accel[0].X = int32(buf[0])<<2 | (int32(buf[3])>>4)&0x03
	accel[0].Y = int32(buf[1])<<2 | (int32(buf[3])>>2)&0x03
	accel[0].Z = int32(buf[2])<<2 | (int32(buf[3])>>0)&0x03

	accel[1].X = int32(buf[4])<<2 | (int32(buf[7])>>4)&0x03
	accel[1].Y = int32(buf[5])<<2 | (int32(buf[7])>>2)&0x03
	accel[1].Z = int32(buf[6])<<2 | (int32(buf[7])>>0)&0x03

	speaker = buf[8] & 0x7f
	motor = buf[8]&0x80 != 0

	return
}
