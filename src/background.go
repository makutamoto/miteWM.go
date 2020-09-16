package main

// #include <cairo/cairo-xlib.h>
import "C"

type Background struct {
	window     C.Window
	surface    *C.cairo_surface_t
	image      *C.cairo_surface_t
	cr         *C.cairo_t
	imageScale C.double
}

var background Background

func setBackground(file string) {
	var rootAttributes C.XWindowAttributes
	C.XGetWindowAttributes(display, rootWindow, &rootAttributes)
	background.window = C.XCreateSimpleWindow(
		display, rootWindow,
		rootAttributes.x, rootAttributes.y, C.uint(rootAttributes.width), C.uint(rootAttributes.height),
		0, 0, C.XBlackPixel(display, 0),
	)
	background.surface = C.cairo_xlib_surface_create(
		display, background.window,
		C.XDefaultVisual(display, C.XDefaultScreen(display)),
		rootAttributes.width, rootAttributes.height,
	)
	background.cr = C.cairo_create(background.surface)
	background.image = C.cairo_image_surface_create_from_png(C.CString(file))

	imageWidth := C.cairo_image_surface_get_width(background.image)
	imageHeight := C.cairo_image_surface_get_height(background.image)

	if rootAttributes.width/imageWidth > rootAttributes.height/imageHeight {
		background.imageScale = C.double(rootAttributes.width) / C.double(imageWidth)
	} else {
		background.imageScale = C.double(rootAttributes.height) / C.double(imageHeight)
	}

	C.cairo_set_antialias(background.cr, C.CAIRO_ANTIALIAS_SUBPIXEL)
	C.XMapWindow(display, background.window)
}

func drawBackground() {
	if background.image == nil {
		return
	}
	C.cairo_save(background.cr)
	C.cairo_scale(background.cr,
		background.imageScale, background.imageScale)
	C.cairo_set_source_surface(background.cr, background.image, 0, 0)
	C.cairo_paint(background.cr)
	C.cairo_restore(background.cr)
	C.cairo_surface_flush(background.surface)
}
