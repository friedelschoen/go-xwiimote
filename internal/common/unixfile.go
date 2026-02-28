package common

import (
	"io"
	"syscall"
)

type UnbufferedFile int

func (fd UnbufferedFile) FD() int { return int(fd) }

func (fd UnbufferedFile) Read(b []byte) (int, error) {
	for {
		n, err := syscall.Read(int(fd), b)
		if err == syscall.EINTR {
			continue
		}
		return n, err
	}
}

func (fd UnbufferedFile) Write(b []byte) (int, error) {
	for {
		n, err := syscall.Write(int(fd), b)
		if err == syscall.EINTR {
			continue
		}
		return n, err
	}
}

func (fd UnbufferedFile) Seek(offset int64, whence int) (int64, error) {
	for {
		off, err := syscall.Seek(int(fd), offset, whence)
		if err == syscall.EINTR {
			continue
		}
		return off, err
	}
}

func (fd UnbufferedFile) ReadAt(buf []byte, off int64) (int, error) {
	if _, err := fd.Seek(off, io.SeekStart); err != nil {
		return 0, err
	}
	return fd.Read(buf)
}

func (fd UnbufferedFile) Close() error {
	return syscall.Close(int(fd))
}

func (fd UnbufferedFile) Ioctl(cmd, ptr uintptr) error {
	for {
		_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), cmd, ptr)
		if err == 0 {
			return nil
		}
		if err == syscall.EINTR {
			continue
		}
		return err
	}
}
