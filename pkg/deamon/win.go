package deamon

import (
	"fmt"
	"golang.org/x/sys/windows"
	"time"
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
	hWnd, err := CreateMainWindow()
	//hWnd, err := GetConsoleWindow()
	if err != nil {
		panic(err)
	}

	icon, err := LoadImage(0, windows.StringToUTF16Ptr("D:/Download/dns.ico"), IMAGE_ICON, 0, 0, LR_DEFAULTSIZE|LR_LOADFROMFILE)
	if err != nil {
		panic(err)
	}

	var data = NOTIFYICONDATA{}
	data.CbSize = uint32(unsafe.Sizeof(data))
	data.UFlags = NIF_ICON | NIF_TIP | NIF_MESSAGE
	data.UCallbackMessage = TrayMsg
	data.HWnd = hWnd
	data.HIcon = icon
	copy(data.SzTip[:], windows.StringToUTF16("DDNS"))
	if _, err := Shell_NotifyIcon(NIM_ADD, &data); err != nil {
		panic(err)
	}

	defer func() {
		if _, err := Shell_NotifyIcon(NIM_DELETE, &data); err != nil {

		}
	}()

	SetWinProc(hWnd, wndProc)
	//ShowWindow(hWnd, SW_SHOW)
	//time.Sleep(2 * time.Second)
	//ShowWindow(hWnd, SW_HIDE)
	//time.Sleep(4 * time.Second)
	ShowWindow(hWnd, SW_SHOW)

	go func() {
		for true {
			fmt.Println(CheckWindowMinimize(hWnd))
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
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
	}()

	time.Sleep(10 * time.Second)
}
