package deamon

import (
	"golang.org/x/sys/windows"
	"unsafe"
)

/* C Win API
//send message to taskbar
BOOL Shell_NotifyIconW(
  DWORD            dwMessage,  //what action
  PNOTIFYICONDATAW lpData      //action's info
);
*/

func Show() {
	hwnd, err := CreateMainWindow()
	if err != nil {
		panic(err)
	}

	icon, err := LoadImage(0, windows.StringToUTF16Ptr("D:/Download/dns.ico"), IMAGE_ICON, 0, 0, LR_DEFAULTSIZE|LR_LOADFROMFILE)
	if err != nil {
		panic(err)
	}

	var data = NOTIFYICONDATA{}
	data.CbSize = uint32(unsafe.Sizeof(data))
	data.UFlags = NIF_ICON
	data.HWnd = hwnd
	data.HIcon = icon
	if _, err := Shell_NotifyIcon(NIM_ADD, &data); err != nil {
		panic(err)
	}

	defer func() {
		if _, err := Shell_NotifyIcon(NIM_DELETE, &data); err != nil {

		}
	}()

	ShowWindow(hwnd, SW_SHOW)

	var msg MSG
	for true {
		r, err := GetMessage(&msg, 0, 0, 0)
		if err != nil {
			panic(err)
		}
		if r == 0 {
			break
		}

		TranslateMessage(&msg)
		DispatchMessage(&msg)
	}
}
