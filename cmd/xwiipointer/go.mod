module github.com/friedelschoen/go-xwiimote/cmd/xwiipointer

go 1.25.3

require (
	github.com/friedelschoen/go-xwiimote v0.0.0
	github.com/friedelschoen/go-xwiimote/pkg/virtpointer v0.0.0
	github.com/friedelschoen/wayland v0.2.0
)

require golang.org/x/sys v0.38.0 // indirect

replace (
	github.com/friedelschoen/go-xwiimote v0.0.0 => ../..
	github.com/friedelschoen/go-xwiimote/pkg/virtpointer v0.0.0 => ../../pkg/virtpointer
)
