package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"unsafe"
)

// #cgo LDFLAGS: -lX11
// #include <X11/Xlib.h>
// #include <X11/cursorfont.h>
// #include <X11/Xutil.h>
import "C"

type GlobalInfo struct {
	button, eventProperty uint
	xRoot, yRoot          int
	window                C.Window
	attributes            C.XWindowAttributes
}

type MiteWMConfig struct {
	BackgroundImageFile string `json:"background_image_file"`
}

const (
	CLIENT_WIDGETS = 2 // 構成ウインドウの数
	CLIENT_BOX     = 0 // Box
	CLIENT_APP     = 1 // Application
	CLIENT_MASK    = 2 // Mask

	RESIZE_ANGLE_TOP    = 0
	RESIZE_ANGLE_BOTTOM = 1
	RESIZE_ANGLE_START  = 2
	RESIZE_ANGLE_END    = 3
	EXIT_PRESSED        = 4

	CONFIG_BOX_BORDER = 12
	// クライアントの枠線の幅のうち、影を描画しない外側の部分。
	CONFIG_EMPTY_BOX_BORDER = 4
	// クライアントのTITLEBARの右端とAPPの右端のX座標の差 = EXITの幅
	CONFIG_TITLEBAR_WIDTH_MARGIN = 25
	// TITLEBARの高さ
	CONFIG_TITLEBAR_HEIGHT = 25
	// クライアントの影の粗さ。必ず0より大きい値に！
	// 1.5~2.5 ぐらいがいい感じ。粗すぎると変な見た目になるし、逆に
	CONFIG_SHADOW_ROUGHNESS = 1.5

	WIDTH_DIFF  = 2 * CONFIG_BOX_BORDER
	HEIGHT_DIFF = CONFIG_TITLEBAR_HEIGHT + 2*CONFIG_BOX_BORDER
)

var (
	display    *C.Display
	rootWindow C.Window
)

func main() {
	var config MiteWMConfig
	if len(os.Args) > 1 {
		log.Println(os.Args[1])
		data, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			log.Println("Failed to load Configuration file.")
		} else {
			err = json.Unmarshal(data, &config)
			if err != nil {
				log.Fatalln("Failed to parse Configuration file.")
			}
		}
	}
	display = C.XOpenDisplay(nil)
	defer C.XCloseDisplay(display)
	if display == nil {
		log.Fatalln("Failed to Open Display.")
	}

	clientTable := make(ClientTable)

	rootWindow = C.XDefaultRootWindow(display)

	C.XSelectInput(
		display, rootWindow,
		C.ButtonPressMask|C.ButtonReleaseMask|C.PointerMotionMask|C.SubstructureNotifyMask,
	)

	if config.BackgroundImageFile != "" {
		setBackground(config.BackgroundImageFile)
	}

	var info GlobalInfo
	info.window = C.None

	var lastUngrabbedApp C.Window = C.None

	C.XDefineCursor(
		display, rootWindow,
		C.XCreateFontCursor(display, C.XC_left_ptr),
	)
	for true {
		var event C.XEvent
		if C.XPending(display) < 0 || C.XNextEvent(display, &event) < 0 {
			break
		}
		switch *(*C.int)(unsafe.Pointer(&event)) {
		case C.MapNotify:
			xmap := (*C.XMapEvent)(unsafe.Pointer(&event))
			if xmap.override_redirect != 0 {
				break
			}
			if xmap.window == background.window {
				drawBackground()
				break
			}
			newClient(&clientTable, xmap.window, &lastUngrabbedApp)
			if lastUngrabbedApp != C.None {
				client := clientTable.findFromApp(lastUngrabbedApp)
				client.drawClient()
			}
			lastUngrabbedApp = C.None
		case C.ButtonPress:
			xbutton := (*C.XButtonEvent)(unsafe.Pointer(&event))
			if xbutton.subwindow == C.None || xbutton.subwindow == background.window {
				break
			}
			client, exist := clientTable[xbutton.subwindow]

			if !exist {
				client = clientTable.findFromApp(xbutton.window)
				if client == nil {
					break
				}
			} else {
				// 掴まれているウインドウの情報を更新する作業。
				C.XGetWindowAttributes(display, client.window[CLIENT_BOX], &info.attributes)

				info.button = uint(xbutton.button)
				info.window = xbutton.subwindow
				info.xRoot = int(xbutton.x_root)
				info.yRoot = int(xbutton.y_root)

				// マウスボタンイベント情報を手に入れる。
				// ここでgrip_info.event_propertyに格納された値は、MotionNotifyの項でも利用する。
				xmotion := (*C.XMotionEvent)(unsafe.Pointer(&event))
				client.setButtonEventInfo(
					int(xmotion.x-info.attributes.x),
					int(xmotion.y-info.attributes.y),
					int(info.attributes.width),
					int(info.attributes.height),
					&info.eventProperty,
				)

				// EXITボタンが押されている時
				if (info.eventProperty>>EXIT_PRESSED)&1 == 1 {

					// APPが「自ら除去イベントを送信した」という状況を再現している。
					// しかし、APPは実は本意ではなかったのかもしれない。
					// 悲しい。

					// 除去イベントを設定
					var deleteEvent C.XEvent
					xclient := (*C.XClientMessageEvent)(unsafe.Pointer(&deleteEvent))
					xclient._type = C.ClientMessage
					xclient.message_type = C.XInternAtom(display, C.CString("WM_PROTOCOLS"), C.True)
					xclient.format = 32
					l := (*[]C.ulong)(unsafe.Pointer(&xclient.data))
					(*l)[0] = C.XInternAtom(display, C.CString("WM_DELETE_WINDOW"), C.True)
					(*l)[1] = C.CurrentTime
					xclient.window = client.window[CLIENT_APP]

					// 除去イベントを送信
					C.XSendEvent(display, client.window[CLIENT_APP], C.False, C.NoEventMask, &deleteEvent)
				}
			}

			C.XRaiseWindow(display, client.window[CLIENT_BOX])
			C.XSetInputFocus(display, client.window[CLIENT_APP], C.RevertToNone, C.CurrentTime)

			if lastUngrabbedApp != C.None {
				client := clientTable.findFromApp(lastUngrabbedApp)
				client.drawClient()
				C.XGrabButton(
					display,
					C.AnyButton,
					C.AnyModifier,
					lastUngrabbedApp,
					C.False,
					C.ButtonPressMask,
					C.GrabModeAsync,
					C.GrabModeAsync,
					C.None,
					C.None,
				)
			}

			if lastUngrabbedApp != client.window[CLIENT_APP] {
				client.drawClient()
			}

			C.XUngrabButton(display, C.AnyButton, C.AnyModifier, client.window[CLIENT_APP])
			lastUngrabbedApp = client.window[CLIENT_APP]
		case C.ButtonRelease:
			// 掴んだウインドウを離す。
			info.window = C.None
			// カーソルを定義。
			C.XDefineCursor(
				display, rootWindow,
				C.XCreateFontCursor(display, C.XC_left_ptr),
			)
		case C.ConfigureNotify:
			xconfigure := (*C.XConfigureEvent)(unsafe.Pointer(&event))
			client := clientTable.findFromApp(xconfigure.window)
			client.configNotify(&event)
		case C.DestroyNotify:
			// 除去イベント。必ずAPPが除去されている時に送信されなければいけない。

			// BOXもろとも除去してしまう。
			xclient := (*C.XClientMessageEvent)(unsafe.Pointer(&event))
			if xclient.window == C.None || xclient.window == rootWindow {
				break
			}

			client, exist := clientTable[xclient.window]

			if !exist {
				break
			}

			if lastUngrabbedApp == client.window[CLIENT_APP] {
				lastUngrabbedApp = C.None
			}
			C.XDestroyWindow(display, client.window[CLIENT_BOX])
			delete(clientTable, client.window[CLIENT_BOX])
		case C.MotionNotify:
			if info.window == C.None {
				break
			}

			client, exist := clientTable[info.window]
			if !exist {
				break
			}

			xbutton := (*C.XButtonEvent)(unsafe.Pointer(&event))
			xDiff := int(xbutton.x_root) - info.xRoot
			yDiff := int(xbutton.y_root) - info.yRoot

			// 掴んでいる部分がウインドウの端でなければ
			if info.eventProperty == 0 {
				// ウインドウを動かす。
				C.XMoveWindow(
					display,
					info.window,
					info.attributes.x+C.int(xDiff),
					info.attributes.y+C.int(yDiff),
				)
			} else {
				// 掴んでいる位置によって、x及びy方向へのresizeが適用されないこともある。
				if (info.eventProperty>>RESIZE_ANGLE_START)&1 == 0 ||
					(info.eventProperty>>RESIZE_ANGLE_END)&1 == 0 {
					xDiff = 0
				}
				if (info.eventProperty>>RESIZE_ANGLE_TOP)&1 == 0 ||
					(info.eventProperty>>RESIZE_ANGLE_BOTTOM)&1 == 0 {
					yDiff = 0
				}

				// カーソル情報。
				cursorInfo := C.XC_left_ptr
				if (info.eventProperty>>RESIZE_ANGLE_TOP)&1 == 1 {
					if (info.eventProperty>>RESIZE_ANGLE_START)&1 == 1 {
						cursorInfo = C.XC_top_left_corner
					} else if (info.eventProperty>>RESIZE_ANGLE_END)&1 == 1 {
						cursorInfo = C.XC_top_right_corner
					} else {
						cursorInfo = C.XC_top_side
					}
				} else if (info.eventProperty>>RESIZE_ANGLE_BOTTOM)&1 == 1 {
					if (info.eventProperty>>RESIZE_ANGLE_START)&1 == 1 {
						cursorInfo = C.XC_bottom_left_corner
					} else if (info.eventProperty>>RESIZE_ANGLE_END)&1 == 1 {
						cursorInfo = C.XC_bottom_right_corner
					} else {
						cursorInfo = C.XC_bottom_side
					}
				} else {
					if (info.eventProperty>>RESIZE_ANGLE_START)&1 == 1 {
						cursorInfo = C.XC_left_side
					} else if (info.eventProperty>>RESIZE_ANGLE_END)&1 == 1 {
						cursorInfo = C.XC_right_side
					}
				}

				C.XDefineCursor(
					display,
					rootWindow,
					C.XCreateFontCursor(display, C.uint(cursorInfo)),
				)

				client.resizeWindow(
					int(info.attributes.x),
					int(info.attributes.y),
					int(info.attributes.width),
					int(info.attributes.height),
					xDiff,
					yDiff,
					(info.eventProperty>>RESIZE_ANGLE_START)&1 == 1,
					(info.eventProperty>>RESIZE_ANGLE_TOP)&1 == 1,
				)
			}
		}
	}
}
