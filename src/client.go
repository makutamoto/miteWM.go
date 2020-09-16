package main

// #include <cairo/cairo-xlib.h>
import "C"

type Client struct {
	window                              [CLIENT_WIDGETS]C.Window
	surface                             [1]*C.cairo_surface_t
	cr                                  [1]*C.cairo_t
	title                               string
	localBorderWidth, localBorderHeight int
}

func (client *Client) drawClient() {
	overallWidth := int(C.cairo_xlib_surface_get_width(client.surface[CLIENT_BOX]))
	overallHeight := int(C.cairo_xlib_surface_get_height(client.surface[CLIENT_BOX]))

	clientWidth := overallWidth - 2*CONFIG_EMPTY_BOX_BORDER
	clientHeight := overallHeight - 2*CONFIG_EMPTY_BOX_BORDER

	shadeBoxBorder := CONFIG_BOX_BORDER - CONFIG_EMPTY_BOX_BORDER + CONFIG_SHADOW_ROUGHNESS

	for i := 0.0; i < shadeBoxBorder; i += CONFIG_SHADOW_ROUGHNESS {
		C.cairo_set_source_rgba(
			client.cr[CLIENT_BOX], 0, 0, 0,
			C.double(0.1/shadeBoxBorder*i*i/CONFIG_SHADOW_ROUGHNESS),
		)
		C.cairo_rectangle(
			client.cr[CLIENT_BOX],
			C.double(i+CONFIG_EMPTY_BOX_BORDER),
			C.double(i+CONFIG_EMPTY_BOX_BORDER),
			C.double(float64(clientWidth)-i*2.0),
			C.double(float64(clientHeight)-i*2),
		)
		C.cairo_stroke(client.cr[CLIENT_BOX])
	}
	titlebarEmphasization := 0.0

	// TITLEBAR描画の前に、ターゲットがFOCUSされているか確認(されている場合、色を強調)
	var focusedWindow C.Window
	var focusedWindowRevert C.int
	C.XGetInputFocus(display, &focusedWindow, &focusedWindowRevert)
	if focusedWindow == client.window[CLIENT_APP] && focusedWindowRevert == C.CurrentTime {
		titlebarEmphasization = 1.0
	}

	titlebarColorA := 0.4 - titlebarEmphasization*0.2
	titlebarColorB := 0.6 - titlebarEmphasization*0.2

	// 描画
	var titlebarWidth = overallWidth - client.localBorderWidth
	titlebarPattern := C.cairo_pattern_create_linear(
		0.0, 0.0, C.double(titlebarWidth*0.0), CONFIG_TITLEBAR_HEIGHT*1.0,
	)

	C.cairo_pattern_add_color_stop_rgb(titlebarPattern, 1, C.double(titlebarColorA), C.double(titlebarColorA), C.double(titlebarColorA))
	C.cairo_pattern_add_color_stop_rgb(titlebarPattern, 0, C.double(titlebarColorB), C.double(titlebarColorB), C.double(titlebarColorB))
	C.cairo_set_source(client.cr[CLIENT_BOX], titlebarPattern)

	C.cairo_rectangle(
		client.cr[CLIENT_BOX],
		C.double(client.localBorderWidth/2),
		C.double(CONFIG_BOX_BORDER),
		C.double(titlebarWidth),
		CONFIG_TITLEBAR_HEIGHT,
	)
	C.cairo_fill(client.cr[CLIENT_BOX])

	C.cairo_set_source_rgb(client.cr[CLIENT_BOX], 0.9, 0.9, 0.9)

	C.cairo_select_font_face(
		client.cr[CLIENT_BOX],
		C.CString("FreeMono"),
		C.CAIRO_FONT_SLANT_NORMAL,
		C.CAIRO_FONT_WEIGHT_BOLD,
	)

	C.cairo_set_font_size(client.cr[CLIENT_BOX], 20)
	C.cairo_move_to(client.cr[CLIENT_BOX], C.double(client.localBorderWidth/2+5.0), CONFIG_TITLEBAR_HEIGHT+CONFIG_BOX_BORDER-2.0)
	C.cairo_show_text(client.cr[CLIENT_BOX], C.CString(client.title))

	// EXITボタンの描画
	exitX := overallWidth - CONFIG_TITLEBAR_WIDTH_MARGIN - client.localBorderWidth/2
	marginWidth := 4
	shadowWidth := 2

	C.cairo_set_source_rgb(client.cr[CLIENT_BOX], 1.0, 0.3, 0.3)
	C.cairo_rectangle(
		client.cr[CLIENT_BOX],
		C.double(exitX+marginWidth),
		C.double(CONFIG_BOX_BORDER+marginWidth),
		C.double(CONFIG_TITLEBAR_WIDTH_MARGIN-marginWidth*2),
		C.double(CONFIG_TITLEBAR_HEIGHT-marginWidth*2),
	)
	C.cairo_fill_preserve(client.cr[CLIENT_BOX])

	C.cairo_set_source_rgb(client.cr[CLIENT_BOX], 0.7, 0.2, 0.2)
	C.cairo_set_line_width(client.cr[CLIENT_BOX], C.double(shadowWidth))
	C.cairo_stroke(client.cr[CLIENT_BOX])

	//cairo_surface_flush(_client->surface[MTWM_CLIENT_BOX]);
}

// カーソルがBOX内のどこに触れているのか？を調べる
func (client *Client) setButtonEventInfo(
	pointerX, pointerY int,
	windowWidth, windowHeight int,
	eventProperty *uint,
) {
	(*eventProperty) = 0

	start := client.localBorderWidth / 2
	top := CONFIG_BOX_BORDER
	end := windowWidth - client.localBorderWidth/2
	bottom := windowHeight - CONFIG_BOX_BORDER

	if pointerX >= end-CONFIG_TITLEBAR_WIDTH_MARGIN &&
		pointerY <= top+CONFIG_TITLEBAR_HEIGHT &&
		pointerX <= end &&
		pointerY >= top {
		(*eventProperty) += 1 << EXIT_PRESSED
		return
	}

	if pointerX < start {
		(*eventProperty) += 1 << RESIZE_ANGLE_START
	}
	if pointerY < top {
		(*eventProperty) += 1 << RESIZE_ANGLE_TOP
	}
	if pointerX > end {
		(*eventProperty) += 1 << RESIZE_ANGLE_END
	}
	if pointerY > bottom {
		(*eventProperty) += 1 << RESIZE_ANGLE_BOTTOM
	}
}
