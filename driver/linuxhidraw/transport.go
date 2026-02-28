package linuxhidraw

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/friedelschoen/go-wiimote"
	"github.com/friedelschoen/go-wiimote/driver/commonhid"
	"github.com/friedelschoen/go-wiimote/internal/common"
)

func NewTransportFromInfo(info wiimote.DeviceInfo) (commonhid.Transport, error) {
	entries, err := os.ReadDir(filepath.Join(info.Syspath(), "hidraw"))
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, os.ErrNotExist
	}
	hidrawpath := filepath.Join("/dev", entries[0].Name())

	//syscall.O_RDWR|
	fd, err := syscall.Open(hidrawpath, syscall.O_RDWR|syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	return common.UnbufferedFile(fd), nil
}
