module github.com/friedelschoen/go-xwiimote/cmd/xwiimap

go 1.25.3

require (
	github.com/bendahl/uinput v1.7.0
	github.com/friedelschoen/go-xwiimote v0.0.0
	golang.org/x/sys v0.37.0
)

replace github.com/friedelschoen/go-xwiimote v0.0.0 => ../..
