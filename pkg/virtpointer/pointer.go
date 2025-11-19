package virtpointer

import (
	"log"
	"time"

	"github.com/friedelschoen/go-xwiimote/pkg/virtpointer/proto"
	"github.com/friedelschoen/wayland"
)

//go:generate gowls -p proto -P wl_ proto/wayland.xml
//go:generate gowls -p proto -P wl_,zwlr_ -S _v1 proto/wlr-virtual-pointer-unstable-v1.xml

type VirtualPointer struct {
	conn    *wayland.Conn
	disp    *proto.Display
	manager *proto.VirtualPointerManager
	pointer *proto.VirtualPointer
}

func NewVirtualPointer() (*VirtualPointer, error) {
	vp := &VirtualPointer{}
	var err error
	vp.conn, err = wayland.Connect("")
	if err != nil {
		return nil, err
	}
	vp.disp = proto.NewDisplay(proto.DisplayHandlers{
		OnDeleteID: wayland.EventHandlerFunc[wayland.Event](vp.conn.UnregisterEvent),
		OnError: wayland.EventHandlerFunc[*proto.DisplayErrorEvent](func(ev *proto.DisplayErrorEvent) bool {
			log.Fatalln(ev)
			return true
		}),
	})
	vp.conn.Register(vp.disp)

	vp.manager = proto.NewVirtualPointerManager()
	vp.disp.GetRegistry(wayland.Registrar(vp.manager))
	wayland.Sync(vp.disp) // issue registry
	wayland.Sync(vp.disp) // binding all interfaces

	if !vp.manager.Valid() {
		log.Fatalf("wayland compositor does not support `%s`\n", vp.manager.Name())
	}

	vp.pointer = vp.manager.CreateVirtualPointer(nil)
	return vp, nil
}

func (i *VirtualPointer) Scroll(horiz bool, value float64) {
	axis := proto.PointerAxisVerticalScroll
	if horiz {
		axis = proto.PointerAxisHorizontalScroll
	}
	i.pointer.Axis(uint32(time.Now().UnixMilli()), axis, value)
	i.pointer.Frame()
	wayland.Sync(i.disp)
}

func (i *VirtualPointer) ScrollDiscrete(horiz bool, value float64, discrete int32) {
	axis := proto.PointerAxisVerticalScroll
	if horiz {
		axis = proto.PointerAxisHorizontalScroll
	}
	i.pointer.AxisDiscrete(uint32(time.Now().UnixMilli()), axis, value, discrete)
	i.pointer.Frame()
	wayland.Sync(i.disp)
}

func (i *VirtualPointer) ScollStop(horiz bool) {
	axis := proto.PointerAxisVerticalScroll
	if horiz {
		axis = proto.PointerAxisHorizontalScroll
	}
	i.pointer.AxisStop(uint32(time.Now().UnixMilli()), axis)
	i.pointer.Frame()
	wayland.Sync(i.disp)
}

func (i *VirtualPointer) Button(button wayland.Button, pressed bool) {
	state := proto.PointerButtonStateReleased
	if pressed {
		state = proto.PointerButtonStatePressed
	}
	i.pointer.Button(uint32(time.Now().UnixMilli()), uint32(button), state)
	wayland.Sync(i.disp)
}

func (i *VirtualPointer) Move(dx float64, dy float64) {
	i.pointer.Motion(uint32(time.Now().UnixMilli()), dx, dy)
	wayland.Sync(i.disp)
}

func (i *VirtualPointer) Set(x uint32, y uint32, xExtent uint32, yExtent uint32) {
	i.pointer.MotionAbsolute(uint32(time.Now().UnixMilli()), x, y, xExtent, yExtent)
	wayland.Sync(i.disp)
}

func (i *VirtualPointer) Commit() {
	i.pointer.Frame()
}

func (i *VirtualPointer) Close() error {
	i.pointer.Destroy()
	i.manager.Destroy()
	i.disp.Destroy()
	return i.conn.Close()
}
