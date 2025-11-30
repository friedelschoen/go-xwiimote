# go-xwiimote - Go bindings for xwiimote

This repository contains bindings for [**libxwiimote**](https://xwiimote.github.io/xwiimote/) and some helpers. Detailed documentation is located [*here*](https://pkg.go.dev/github.com/friedelschoen/go-xwiimote).

```
xwiimote            -- bindings for libxwiimote
├── pkg
│   ├── irpointer   -- algorithm to convert IR events to a pointer on a screen
│   └── vinput      -- library to create a virtual input device using Linux' uinput
└── cmd
    ├── xwiimap     -- utility to map wiimote buttons to physical keys.
    └── xwiipointer -- utility to use wiimote as mouse using IR-tracking.
```

*libxwiimote* is a library which cooperates with the [*xwiimote*-kernel driver](https://www.bluez.org/gsoc-nintendo-wii-remote-device-driver/) which is included since Linux 3.1 and supersedes cwiid which is a driverless interface.

## Prerequisite

- Linux 3.1 or newer (3.11 or newer recommended)
- bluez 4.101 or newer (bluez-5.0 or newer recommended)
- [Go](https://go.dev/)
- [libxwiimote](https://github.com/xwiimote/xwiimote)

Supported devices are:

- Nintendo Wii Remote (Nintendo RVL-CNT-01)
- Nintendo Wii Remote Plus (Nintendo RVL-CNT-01-TR)
- Nintendo Wii Nunchuk Extension
- Nintendo Wii Classic Controller Extension
- Nintendo Wii Classic Controller Pro Extension
- Nintendo Wii Balance Board (Nintendo RVL-WBC-01)
- Nintendo Wii U Pro Controller (Nintendo RVL-CNT-01-UC)
- Nintendo Wii Guitar Extensions
- Nintendo Wii Drums Extensions

## Usage

For example usage you can check the `cmd`-directory.

### Find new devices

First you have to choose whether you want to monitor for new devices or only use currently available devices. If you want to monitor for new devices you should use `Montor`:
```go
monitor := xwiimote.NewMonitor(xwiimote.MonitorUdev)
defer monitor.Free()

for {
    // Wait infinitely for a new device.
    path, err := monitor.Wait(-1)
    if err != nil || path == "" {
        log.Printf("error while polling: %v\n", err)
        continue
    }
    -> device at path
}
```

If you only want to use currently available devices, you can use the `IterDevices`-function:
```go
for path := range xwiimote.IterDevices(xwiimote.MonitorUdev) {
    -> device at path
}
```

The path returned points at the sysfs location.

### Create a device

This is a sparse example how to create a new device. Refer to the documentation for more information.
```go
// create a new device which is located at path
dev, err := xwiimote.NewDevice(path)
if err != nil {
    log.Fatalf("error: unable to get device: %s", err)
}
// freeing is not mandatory and is done automatically by GC.
defer dev.Free()

// open interfaces, we're only interested in core functionality.
if err := dev.Open(xwiimote.InterfaceCore); err != nil {
    log.Fatalf("error: unable to open device: %s", err)
}

for {
    // Wait infinitely for a new event.
    ev, err := dev.Wait(-1)
    if err != nil {
        log.Printf("unable to poll event: %v\n", err)
    }
    switch ev := ev.(type) {
    case *xwiimote.EventKey:
        log.Printf("key event: %v\n", ev.Code)
    }
}
```

### Use a IR pointer

The IRPointer has a state which must be updated when appropriate, after updating the health and position can be read.
```go
pointer := irpointer.NewIRPointer(nil)
var (
    lastIR *xwiimote.EventIR
    lastAccel *xwiimote.EventAccel
)
for {
    ev, err := dev.Wait(-1)
    if err != nil {
        log.Printf("unable to poll event: %v\n", err)
    }
    switch ev := ev.(type) {
    case *xwiimote.EventIR:
        lastIR = ev
    case *xwiimote.EventAccel:
        lastAccel = ev
    if lastIR != nil && lastAccel != nil {
        pointer.Update(lastIR.Slots, lastAccel.Accel)
        /* optionally only update when there is a new IR AND Accel event
        lastIR = nil
        lastAccel = nil
        */
    }
    // if the pointer has sufficient health and a valid position -> do somthing
    if pointer.Health >= irpointer.IRSingle && pointer.Position != nil {
        x, y := pointer.Position.X, pointer.Position.Y
        if x >= -340 && x < 340 && y >= -92 && y < 290 {
            fmt.Printf("[%v] pointer at (%.2f %.2f) at %.2fm distance\n", pointer.Health, pointer.Position.X, pointer.Position.Y, pointer.Distance)
        }
    }
}
```

## Contributing

Feel free to add functionality and make a pull request! Please keep the structure clean:

```
/       -> only releated to libxwiimote itself and bindings
/pkg    -> utility packages not related to libxwiimote
```

### Tooling

This project makes use of `stringer` to generate `.String()` methods for enums.
- Install `stringer`:
  ```
  $ go install golang.org/x/tools/cmd/stringer
  ```
- Run generators:
  ```
  $ go generate ./...
  ```

### Formatting

Before submitting any changes, please format the project using `gofmt`.

## Licensing

The irpointer package is licensed under 2-clause-BSD License (as noted in the source), remaining code is licensed under Zlib License.