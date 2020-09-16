package main

// #cgo LDFLAGS: -lX11 -lcairo
// #include <X11/Xlib.h>
// #include <X11/cursorfont.h>
// #include <X11/Xutil.h>
// #include <cairo/cairo-xlib.h>
import "C"

type ClientTable map[C.Window]*Client

func (clientTable *ClientTable) findFromApp(window C.Window) *Client {
	var parent, root C.Window
	var child *C.Window
	var childNum C.uint
	C.XQueryTree(display, window, &root, &parent, &child, &childNum)
	return (*clientTable)[parent]
}
