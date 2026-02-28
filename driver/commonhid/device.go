package commonhid

import (
	"errors"
	"io"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/friedelschoen/go-wiimote"
	"github.com/friedelschoen/go-wiimote/internal/common"
)

type commonEvent struct {
	iface     wiimote.Feature
	timestamp time.Time
}

func (e commonEvent) Feature() wiimote.Feature { return e.iface }
func (e commonEvent) Timestamp() time.Time     { return e.timestamp }

type feature struct {
	kind wiimote.FeatureKind
	dev  *device
}

func (f feature) Kind() wiimote.FeatureKind { return f.kind }
func (f feature) Device() wiimote.Device    { return f.dev }
func (f feature) Opened() bool              { return f.dev.openIfs&f.kind != 0 }
func (f feature) Close() error {
	f.dev.openIfs &^= f.kind
	return f.dev.updateReportMode()
}

type memResponse struct {
	err  error
	addr uint16
	data []byte
}

type device struct {
	wiimote.Poller[wiimote.Event]

	transport Transport

	openIfs wiimote.FeatureKind
	// availIfs wiimote.FeatureKind

	// last known button state (16-bit)
	btnPrev uint16

	// last known battery state
	battery uint8

	// wether an extension is currently connected
	hasExtension bool

	// led state
	led    wiimote.Led
	rumble bool
	irfull bool

	memory chan memResponse

	ackErr map[uint8]error
	ackMu  sync.Mutex

	interleaved [19]byte

	// queued events (because a single report can generate multiple key events)
	moreEvents chan wiimote.Event
}

type memory struct {
	dev *device
}

func (m memory) Close() error {
	return nil
}

func (m memory) WriteAt(p []byte, off int64) (n int, err error) {
	return m.dev.writeMemory(memEEPROM, uint32(off), p)
}

func (m memory) ReadAt(p []byte, off int64) (n int, err error) {
	return m.dev.readMemory(memEEPROM, uint32(off), p)
}

type Transport interface {
	io.ReadWriteCloser
	FD() int
}

func NewDevice(transport Transport) wiimote.Device {
	d := &device{
		transport:  transport,
		openIfs:    wiimote.FeatureCore,
		moreEvents: make(chan wiimote.Event, 64),
		memory:     make(chan memResponse, 16),
		ackErr:     make(map[uint8]error),
	}
	d.Poller = common.NewPoller(d)
	return d
}

// Feature interface
func (d *device) Kind() wiimote.FeatureKind { return wiimote.FeatureCore }
func (d *device) Device() wiimote.Device    { return d }
func (d *device) Opened() bool              { return true }
func (d *device) Close() error              { return nil }

func (d *device) Rumble(state bool) error {
	d.rumble = state
	return d.output(true, 0x10, 0x00)
}

func (d *device) Memory() (wiimote.Memory, error) {
	return memory{dev: d}, nil
}

// Device interface
func (d *device) IRFull() bool { return d.irfull }

func (d *device) SetIRFull(fullreport bool) { d.irfull = fullreport }

func (d *device) Syspath() string { return "" }

func (d *device) String() string {
	return "wiimote-device (commonhid)"
}

func (d *device) OpenFeatures(ifaces wiimote.FeatureKind, wr bool) error {
	_ = wr // all features are writeable

	d.openIfs |= ifaces
	return d.updateReportMode()
}

func (d *device) FD() int {
	return d.transport.FD()
}

func (d *device) Feature(kind wiimote.FeatureKind) wiimote.Feature {
	if d.openIfs&kind == 0 {
		return nil
	}
	if kind == wiimote.FeatureCore {
		return d
	}
	return feature{kind: kind, dev: d}
}

func (d *device) Available(iface wiimote.FeatureKind) bool {
	const nonExtension = wiimote.FeatureCore | wiimote.FeatureAccel | wiimote.FeatureIR | wiimote.FeatureSpeaker
	if iface&nonExtension != 0 {
		return true
	}
	// TODO: if extension
	// TODO: if balanceboard ...
	return false
}

func (d *device) LED() (wiimote.Led, error) {
	return d.led, nil
}

func (d *device) SetLED(leds wiimote.Led) error {
	return d.output(true, 0x11, byte(leds)<<4)
}

func (d *device) Battery() (uint, error) {
	return uint(d.battery), nil
}

func (d *device) DevType() (string, error) {
	return "unknown", os.ErrInvalid
}

func (d *device) Extension() (string, error) {
	return "none", os.ErrInvalid
}

func (d *device) Poll() (wiimote.Event, bool, error) {
	select {
	case ev := <-d.moreEvents:
		return ev, true, nil
	default:
	}

	d.readEvent()

	select {
	case ev := <-d.moreEvents:
		return ev, true, nil
	default:
		return nil, false, common.ErrPollAgain
	}
}

func (d *device) readEvent() {
	d.WaitReadable(-1)

	var buf [64]byte
	n, err := d.transport.Read(buf[:])
	if err != nil {
		// Non-blocking transport support (os.File O_NONBLOCK typically returns EAGAIN)
		if errors.Is(err, syscall.EAGAIN) {
			return
		}

		// Device gone
		if errors.Is(err, io.EOF) {
			d.moreEvents <- &wiimote.EventGone{
				Event: commonEvent{iface: nil, timestamp: time.Now()},
			}
			return
		}

		// Any other error: treat like "watch"/state changed; caller should reopen etc.
		d.moreEvents <- &wiimote.EventWatch{
			Event: commonEvent{iface: nil, timestamp: time.Now()},
		}
		return
	}
	if n <= 0 {
		return
	}

	report := buf[:n]
	ts := time.Now()

	rid := report[0]

	if rid == 0x20 {
		d.battery = report[6]
		if report[3]&0x01 != 0 {
			d.battery = 0
		}
		// throw event ?
		d.hasExtension = report[3]&0x02 != 0
		d.led = wiimote.Led(report[3] >> 4)
	}
	if rid == 0x21 {
		size := (report[3] >> 4) + 1
		addr := uint16(report[4])<<8 | uint16(report[5])
		var err error
		switch report[3] & 0x0f {
		case 0:
			err = nil
		case 7:
			err = os.ErrPermission // <---- HERE
		case 8:
			err = io.EOF
		default:
			err = os.ErrInvalid
		}
		d.memory <- memResponse{
			err:  err,
			addr: addr,
			data: report[6 : 6+size],
		}
	}
	if rid == 0x22 {
		ack := report[3]
		var err error
		if report[4] != 0 {
			err = os.ErrInvalid
		}
		d.ackMu.Lock()
		d.ackErr[ack] = err
		d.ackMu.Unlock()
	}
	if rid == 0x20 || rid == 0x21 || rid == 0x22 || rid == 0x30 || rid == 0x31 || rid == 0x33 || rid == 0x35 || rid == 0x37 || rid == 0x3e || rid == 0x3f {
		btn := uint16(report[1])<<8 | uint16(report[2])
		d.emitButtons(ts, btn)
	}
	if rid == 0x31 || rid == 0x33 || rid == 0x35 || rid == 0x37 {
		var accel wiimote.Vec3
		accel.X = (int32(report[3])<<2 | int32(report[1]>>5)&0x03) - 0x80
		accel.Y = (int32(report[4])<<2 | int32(report[1]>>4)&0x03) - 0x80
		accel.Z = (int32(report[5])<<2 | int32(report[1]>>5)&0x02) - 0x80

		d.moreEvents <- &wiimote.EventAccel{
			Event: commonEvent{iface: d, timestamp: ts},
			Accel: accel,
		}
	}
	if rid == 0x33 {
		// extended
		d.emitIRExtended(ts, report[6:18])
	}
	if rid == 0x36 {
		// basic
		d.emitIRBasic(ts, report[3:13])
	}
	if rid == 0x37 {
		// basic
		d.emitIRBasic(ts, report[6:16])
	}
	if rid == 0x3e {
		copy(d.interleaved[:], report[4:])
	}
	if rid == 0x3f {
		var accel wiimote.Vec3
		accel.X = int32(d.interleaved[3])
		accel.Y = int32(report[3])
		accel.Z = (int32(d.interleaved[0]>>5)&0x03)<<6 |
			(int32(d.interleaved[1]>>5)&0x03)<<8 |
			(int32(report[0]>>5)&0x03)<<2 |
			(int32(report[1]>>5)&0x03)<<4

		d.moreEvents <- &wiimote.EventAccel{
			Event: commonEvent{iface: d, timestamp: ts},
			Accel: accel,
		}
		// TODO: IR event
	}

	// TODO: extensions
}

func (d *device) emitButtons(ts time.Time, btn uint16) {
	prev := d.btnPrev
	if btn == prev {
		return
	}
	d.btnPrev = btn

	changed := btn ^ prev
	emitChanged := func(k wiimote.Key, value uint16) {
		if changed&value == 0 {
			return
		}
		pressed := btn&value != 0
		d.moreEvents <- &wiimote.EventKey{
			Event:   commonEvent{iface: d, timestamp: ts},
			Code:    k,
			Pressed: pressed,
		}
	}

	emitChanged(wiimote.KeyTwo, 0x0001)
	emitChanged(wiimote.KeyOne, 0x0002)
	emitChanged(wiimote.KeyB, 0x0004)
	emitChanged(wiimote.KeyA, 0x0008)
	emitChanged(wiimote.KeyMinus, 0x0010)
	emitChanged(wiimote.KeyHome, 0x0080)
	emitChanged(wiimote.KeyLeft, 0x0100)
	emitChanged(wiimote.KeyRight, 0x0200)
	emitChanged(wiimote.KeyDown, 0x0400)
	emitChanged(wiimote.KeyUp, 0x0800)
	emitChanged(wiimote.KeyPlus, 0x1000)
}

func (d *device) emitIRBasic(ts time.Time, report []byte) {
	var slots [4]wiimote.IRSlot
	slots[0].X = int32(report[0]) | (int32(report[2]>>4)&0x03)<<8
	slots[0].Y = int32(report[1]) | (int32(report[2]>>6)&0x03)<<8
	slots[1].X = int32(report[3]) | (int32(report[2]>>0)&0x03)<<8
	slots[1].Y = int32(report[4]) | (int32(report[2]>>2)&0x03)<<8

	slots[2].X = int32(report[5]) | (int32(report[7]>>4)&0x03)<<8
	slots[2].Y = int32(report[6]) | (int32(report[7]>>6)&0x03)<<8
	slots[3].X = int32(report[8]) | (int32(report[7]>>0)&0x03)<<8
	slots[3].Y = int32(report[9]) | (int32(report[7]>>2)&0x03)<<8

	d.moreEvents <- &wiimote.EventIR{
		Event: commonEvent{iface: d, timestamp: ts},
		Slots: slots,
	}
}

func (d *device) emitIRExtended(ts time.Time, report []byte) {
	var slots [4]wiimote.IRSlot
	slots[0].X = int32(report[0]) | (int32(report[2]>>4)&0x03)<<8
	slots[0].Y = int32(report[1]) | (int32(report[2]>>6)&0x03)<<8
	slots[0].Size = report[2] & 0x0f

	d.moreEvents <- &wiimote.EventIR{
		Event: commonEvent{iface: d, timestamp: ts},
		Slots: slots,
	}
}

func (d *device) output(ack bool, report ...byte) error {
	if len(report) < 2 {
		panic("invalid report")
	}

	rid := report[0]

	if ack {
		// set 0x02 to enable acknowlegding
		report[1] |= 0x02
	}
	if d.rumble {
		report[1] |= 0x01
	}

	if ack {
		d.ackMu.Lock()
		delete(d.ackErr, rid)
		d.ackMu.Unlock()
	}

	if _, err := d.transport.Write(report); err != nil {
		return err
	}

	if !ack {
		return nil
	}

	for {
		d.ackMu.Lock()
		if err, ok := d.ackErr[rid]; ok {
			delete(d.ackErr, rid)
			d.ackMu.Unlock()
			return err
		}
		d.ackMu.Unlock()

		d.readEvent()
	}
}
func (d *device) enable(ack bool, report byte, enable bool) error {
	var enbyte byte
	if enable {
		enbyte = 0x04
	}
	return d.output(ack, report, enbyte)
}

func (d *device) updateReportMode() error {
	if d.openIfs == 0 {
		// do nothing
		return nil
	}
	extension := d.openIfs&^wiimote.FeatureSetCore != 0

	has := func(f wiimote.FeatureKind) bool {
		return d.openIfs&f == f
	}

	var mode byte
	switch {
	// core, accel, ir, extension
	case extension && has(wiimote.FeatureAccel|wiimote.FeatureIR):
		mode = 0x37
	// core, accel, extension
	case extension && has(wiimote.FeatureAccel):
		mode = 0x35
	// core, ir, extension
	case extension && has(wiimote.FeatureIR):
		mode = 0x36
	// core, extension
	case extension && has(wiimote.FeatureCore):
		mode = 0x34
	// GEEN core, alleen extension
	case extension:
		mode = 0x3d
	case has(wiimote.FeatureIR): // and accel but thats reported anyway
		if d.irfull {
			mode = 0x3e
		} else {
			mode = 0x33
		}
	case has(wiimote.FeatureAccel):
		mode = 0x31
	default:
		mode = 0x30
	}

	d.output(true, 0x20, 0x00, mode)
	d.enable(true, 0x13, has(wiimote.FeatureIR))
	d.enable(true, 0x1a, has(wiimote.FeatureIR))
	return nil
}

/* memory spaces */
type memSpace uint8

const (
	memEEPROM memSpace = 0
	memREG    memSpace = 1
)

/*
readMemory reads from EEPROM or register space.
addr is 16-bit for EEPROM and 24-bit for regs (you pass it as uint32 either way).
Returns exactly len(dst) bytes on success.
*/
func (d *device) readMemory(space memSpace, addr uint32, dst []byte) (int, error) {
	if len(dst) == 0 {
		return 0, nil
	}

	off := 0
	for off < len(dst) {
		/* conservative chunk:
		   Many stacks use up to 16 bytes per reply, sometimes 32.
		   We'll request 16 to be safe across transports. */
		n := min(len(dst)-off, 16)

		/* send read request: 0x17 */
		ah, am, al := encodeAddr(space, addr+uint32(off))
		sizeHi, sizeLo := byte(n>>8), byte(n)
		if err := d.output(false, 0x17, 0x00, ah, am, al, sizeHi, sizeLo); err != nil {
			return off, err
		}

		var resp *memResponse
		for resp == nil {
			d.readEvent()
			select {
			case r := <-d.memory:
				resp = &r
			default:
			}
		}
		if resp.err != nil {
			return off, resp.err
		}
		gotAddr := resp.addr
		data := resp.data

		/* validate it’s the reply we asked for */
		if gotAddr != uint16((addr+uint32(off))&0xffff) {
			/* not fatal per se (could be another outstanding read),
			   but in this design we do synchronous single-flight reads,
			   so treat as invalid */
			return off, os.ErrInvalid
		}
		if len(data) < n {
			return off, io.ErrUnexpectedEOF
		}

		copy(dst[off:off+n], data[:n])
		off += n
	}

	return off, nil
}

/*
writeMemory writes to EEPROM or register space.
It chunks writes (max 16 bytes per write report payload here).
For EEPROM writes you often need delays; we keep it simple: sync + small sleep.
*/
func (d *device) writeMemory(space memSpace, addr uint32, src []byte) (int, error) {
	if len(src) == 0 {
		return 0, nil
	}

	off := 0
	for off < len(src) {
		n := len(src) - off
		if n > 16 {
			n = 16
		}

		ah, am, al := encodeAddr(space, addr+uint32(off))

		/* output report 0x16: [16][flags][addr_hi][addr_mid][addr_lo][size][data...] */
		rep := make([]byte, 0, 6+n)
		rep = append(rep, 0x16, 0x00, ah, am, al, byte(n))
		rep = append(rep, src[off:off+n]...)

		if err := d.output(false, rep...); err != nil {
			return off, err
		}

		/* EEPROM writes are slow; typical practice is small delay.
		   For registers it’s usually fine but harmless. */
		time.Sleep(10 * time.Millisecond)

		off += n
	}

	return off, nil
}

/* helper: encode address with space bit */
func encodeAddr(space memSpace, addr uint32) (hi, mid, lo byte) {
	/*
	   Wiimote convention: highest bit of hi-byte selects space:
	   - EEPROM: 0
	   - REG:    1
	   And remaining 23 bits are address.
	*/
	a := addr & 0x00ffffff
	hi = byte((a >> 16) & 0xff)
	if space == memREG {
		hi |= 0x10 /* NOTE: some docs say bit4 selects space; others use bit7.
		   If your reads fail, flip this to 0x80.
		   (I’m choosing 0x10 because that’s common in Wiibrew snippets.) */
	}
	mid = byte((a >> 8) & 0xff)
	lo = byte(a & 0xff)
	return
}
