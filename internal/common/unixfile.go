package common

import (
	"io"
	"syscall"
)

type UnbufferedFile int

func (fd UnbufferedFile) FD() int {
	return int(fd)
}

func (fd UnbufferedFile) Read(b []byte) (int, error) {
	return syscall.Read(int(fd), b)
}

func (fd UnbufferedFile) Write(b []byte) (int, error) {
	return syscall.Write(int(fd), b)
}

func (fd UnbufferedFile) Seek(offset int64, whence int) (int64, error) {
	return syscall.Seek(int(fd), offset, whence)
}

func (fd UnbufferedFile) ReadAt(buf []byte, off int64) (n int, err error) {
	if _, err := fd.Seek(off, io.SeekStart); err != nil {
		return 0, err
	}
	return fd.Read(buf)
}

func (fd UnbufferedFile) Close() error {
	return syscall.Close(int(fd))
}

func (fd UnbufferedFile) Ioctl(cmd, ptr uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), cmd, ptr)
	if err == 0 {
		return nil
	}
	return err
}
