package main

// #include <X11/Xutil.h>
// #include <cairo/cairo-xlib.h>
import "C"
import "unsafe"

func (client *Client) resizeWindow(
	x, y int,
	width, height int,
	xDiff, yDiff int,
	xMove, yMove bool,
) {
	if C.XEventsQueued(display, C.QueuedAlready) >= 2 {
		return
	}

	if xMove {
		xDiff *= -1
	}
	if yMove {
		yDiff *= -1
	}

	var hints C.XSizeHints
	client.getSizeHints(&hints)

	fWidth := width + xDiff
	fHeight := height + yDiff
	fWidthMin := client.localBorderWidth + 1 + int(hints.min_width)
	fHeightMin := client.localBorderHeight + 1 + int(hints.min_height)

	if fWidth <= fWidthMin {
		xDiff += fWidthMin - fWidth
		fWidth = fWidthMin
	}

	if fHeight <= fHeightMin {
		yDiff += fHeightMin - fHeight
		fHeight = fHeightMin
	}

	posX := x
	if xMove {
		posX -= xDiff
	} else {
		posX += 0
	}
	posY := y
	if yMove {
		posY -= yDiff
	} else {
		posY += 0
	}
	C.XMoveWindow(
		display,
		client.window[CLIENT_BOX],
		C.int(posX),
		C.int(posY),
	)

	C.XResizeWindow(
		display,
		client.window[CLIENT_APP],
		C.uint(fWidth-client.localBorderWidth),
		C.uint(fHeight-client.localBorderHeight),
	)
}

func (client *Client) configNotify(event *C.XEvent) {
	if client == nil {
		return
	}
	xconfigure := (*C.XConfigureEvent)(unsafe.Pointer(event))
	fWidth := int(xconfigure.width) + client.localBorderWidth
	fHeight := int(xconfigure.height) + client.localBorderHeight

	C.XResizeWindow(
		display,
		client.window[CLIENT_BOX],
		C.uint(fWidth),
		C.uint(fHeight),
	)

	C.cairo_xlib_surface_set_size(
		client.surface[CLIENT_BOX],
		C.int(fWidth),
		C.int(fHeight),
	)

	// 描画を更新。
	client.drawClient()
}
