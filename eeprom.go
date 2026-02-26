package wiimote

import (
	"fmt"
	"io"
	"syscall"
)

type Memory struct {
	SysFile
}

func (m Memory) Read(buf []byte) (int, error) {
	n := 0
	for n < len(buf) {
		k, err := m.SysFile.Read(buf[n:])
		n += k
		if err != nil && err != syscall.EINTR {
			if err == io.EOF && n > 0 {
				return n, nil
			} else if err == syscall.EIO {
				if _, serr := m.Seek(1, io.SeekCurrent); serr != nil {
					return 0, fmt.Errorf("%w (unable to skip: %w)", err, serr)
				}
			} else {
				return n, err
			}
		}
		if k == 0 {
			// defensief: voorkom infinite loop
			return n, io.ErrNoProgress
		}
	}
	return n, nil

}

func (f Memory) ReadByte() (byte, error) {
	var buf [1]byte
	n, err := f.SysFile.Read(buf[:])
	if err != nil {
		if _, serr := f.Seek(1, io.SeekCurrent); serr != nil {
			return 0, fmt.Errorf("%w (unable to skip: %w)", err, serr)
		}
		/* silently ignoring errors */
		return 0, nil
	}
	if n == 0 {
		return 0, io.EOF
	}
	return buf[0], nil
}

func (m Memory) ReadAt(buf []byte, off int64) (int, error) {
	if _, err := m.Seek(off, io.SeekStart); err != nil {
		return 0, err
	}
	return m.Read(buf)
}
