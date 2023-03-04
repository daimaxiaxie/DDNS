package deamon

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

const (
	IMAGE_ICON      = 1
	LR_DEFAULTSIZE  = 0x00000040
	LR_LOADFROMFILE = 0x00000010

	SW_SHOW = 5

	CW_USEDEFAULT = ^0x7fffffff
)

const (
	WS_CAPTION          = 0x00c00000
	WS_MAXIMIZEBOX      = 0x00010000
	WS_MINIMIZEBOX      = 0x00020000
	WS_OVERLAPPED       = 0x00000000
	WS_SYSMENU          = 0x00080000
	WS_THICKFRAME       = 0x00040000
	WS_OVERLAPPEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX

	WM_DESTROY       = 0x0002
	WM_SETICON       = 0x0080
	WM_MOUSEMOVE     = 0x0200
	WM_LBUTTONDOWN   = 0x0201
	WM_LBUTTONUP     = 0x0202
	WM_LBUTTONDBLCLK = 0x0203
	WM_RBUTTONDOWN   = 0x0204
	WM_RBUTTONUP     = 0x0205
	WM_RBUTTONDBLCLK = 0x0206
	WM_APP           = 0x8000
)

const (
	NIM_ADD        = 0x00000000
	NIM_MODIFY     = 0x00000001
	NIM_DELETE     = 0x00000002
	NIM_SETFOCUS   = 0x00000003
	NIM_SETVERSION = 0x00000004

	NIF_MESSAGE  = 0x00000001
	NIF_ICON     = 0x00000002
	NIF_TIP      = 0x00000004
	NIF_STATE    = 0x00000008
	NIF_INFO     = 0x00000010
	NIF_GUID     = 0x00000020
	NIF_REALTIME = 0x00000040
	NIF_SHOWTIP  = 0x00000080

	NIS_HIDDEN     = 0x00000001
	NIS_SHAREDICON = 0x00000002

	NIIF_NONE               = 0x00000000
	NIIF_INFO               = 0x00000001
	NIIF_WARNING            = 0x00000002
	NIIF_ERROR              = 0x00000003
	NIIF_USER               = 0x00000004
	NIIF_NOSOUND            = 0x00000010
	NIIF_LARGE_ICON         = 0x00000020
	NIIF_RESPECT_QUIET_TIME = 0x00000080
	NIIF_ICON_MASK          = 0x0000000F

	NIN_BALLOONSHOW      = 0x0402
	NIN_BALLOONTIMEOUT   = 0x0404
	NIN_BALLOONUSERCLICK = 0x0405
)

type NOTIFYICONDATA struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GUIDItem         GUID
	HBalloonIcon     uintptr
}

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type POINT struct {
	X int32
	Y int32
}

type MSG struct {
	Hwnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       POINT
	LPrivate uint32
}

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

var (
	libshell32  = windows.NewLazySystemDLL("shell32.dll")
	libuser32   = windows.NewLazySystemDLL("user32.dll")
	libkernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procShell_NotifyIconW = libshell32.NewProc("Shell_NotifyIconW")
	procLoadImageW        = libuser32.NewProc("LoadImageW")

	procRegisterClassExW = libuser32.NewProc("RegisterClassExW")
	procGetModuleHandleW = libkernel32.NewProc("GetModuleHandleW")
	procCreateWindowExW  = libuser32.NewProc("CreateWindowExW")

	procDefWindowProcW  = libuser32.NewProc("DefWindowProcW")
	procPostQuitMessage = libuser32.NewProc("PostQuitMessage")

	procTranslateMessage = libuser32.NewProc("TranslateMessage")
	procDispatchMessageW = libuser32.NewProc("DispatchMessageW")

	procGetMessageW = libuser32.NewProc("GetMessageW")
	procShowWindow  = libuser32.NewProc("ShowWindow")

	getConsoleTitle  = libkernel32.NewProc("GetConsoleTitleW")
	getConsoleWindow = libkernel32.NewProc("GetConsoleWindow")
	setConsoleTitle  = libkernel32.NewProc("SetConsoleTitleW")
)

func Shell_NotifyIcon(dwMessage uint32, lpData *NOTIFYICONDATA) (int32, error) {
	r, _, err := procShell_NotifyIconW.Call(uintptr(dwMessage), uintptr(unsafe.Pointer(lpData)))
	if r == 0 {
		return 0, err
	}
	return int32(r), nil
}

func LoadImage(hInst uintptr, name *uint16, type_ uint32, cx, cy int32, fuLoad uint32) (uintptr, error) {
	r, _, err := procLoadImageW.Call(hInst, uintptr(unsafe.Pointer(name)), uintptr(type_), uintptr(cx), uintptr(cy), uintptr(fuLoad))
	if r == 0 {
		return 0, err
	}
	return r, nil
}

func GetConsoleTitle(lpConsoleTitle *uint16, nSize uint32) uint32 {
	ret, _, _ := syscall.SyscallN(getConsoleTitle.Addr(), 2, uintptr(unsafe.Pointer(lpConsoleTitle)), uintptr(nSize), 0)

	return uint32(ret)
}

func SetConsoleTitle(lpConsoleTitle *uint16, nSize uint32) error {
	_, _, err := syscall.SyscallN(setConsoleTitle.Addr(), 1, uintptr(unsafe.Pointer(lpConsoleTitle)), 0, 0)
	if err != 0 {
		return err
	}

	return nil
}

func GetModuleHandle(lpModuleName *uint16) (uintptr, error) {
	r, _, err := procGetModuleHandleW.Call(uintptr(unsafe.Pointer(lpModuleName)))
	if r == 0 {
		return 0, err
	}
	return r, nil
}

func RegisterClassEx(Arg1 *WNDCLASSEX) (uint16, error) {
	r, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(Arg1)))
	if r == 0 {
		return 0, err
	}
	return uint16(r), nil
}

func CreateWindowEx(dwExStyle uint32, lpClassName, lpWindowName *uint16, dwStyle uint32, X, Y, nWidth, nHeight int32, hWndParent, HMenu, hInstance uintptr, lpParam unsafe.Pointer) (uintptr, error) {
	r, _, err := procCreateWindowExW.Call(uintptr(dwExStyle), uintptr(unsafe.Pointer(lpClassName)), uintptr(unsafe.Pointer(lpWindowName)), uintptr(dwStyle), uintptr(X), uintptr(Y), uintptr(nWidth), uintptr(nHeight), hWndParent, HMenu, hInstance, uintptr(lpParam))
	if r == 0 {
		return 0, err
	}
	return r, nil
}

func DefWindowProc(hWnd uintptr, Msg uint32, wParam, lParam uintptr) (uintptr, error) {
	r, _, _ := procDefWindowProcW.Call(hWnd, uintptr(Msg), wParam, lParam)
	return r, nil
}

func GetMessage(lpMsg *MSG, hWnd uintptr, uMsgFilterMin, uMsgFilterMax uint32) (int32, error) {
	r, _, err := procGetMessageW.Call(uintptr(unsafe.Pointer(lpMsg)), hWnd, uintptr(uMsgFilterMin), uintptr(uMsgFilterMax))
	if int32(r) == -1 {
		return 0, err
	}
	return int32(r), nil
}

func PostQuitMessage(nExitCode int32) {
	procPostQuitMessage.Call(uintptr(nExitCode))
}

func TranslateMessage(lpMsg *MSG) (int32, error) {
	r, _, _ := procTranslateMessage.Call(uintptr(unsafe.Pointer(lpMsg)))
	return int32(r), nil
}

func DispatchMessage(lpMsg *MSG) (uintptr, error) {
	r, _, _ := procDispatchMessageW.Call(uintptr(unsafe.Pointer(lpMsg)))
	return r, nil
}

func wndProc(hWnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_DESTROY:
		PostQuitMessage(0)
	default:
		r, _ := DefWindowProc(hWnd, msg, wParam, lParam)
		return r
	}
	return 0
}

func CreateTray() {
	icon, err := LoadImage(0, windows.StringToUTF16Ptr("D:/Download/dns.ico"), IMAGE_ICON, 0, 0, LR_DEFAULTSIZE|LR_LOADFROMFILE)
	if err != nil {
		panic(err)
	}

	var data = NOTIFYICONDATA{}
	data.CbSize = uint32(unsafe.Sizeof(data))
	data.HIcon = icon
	if _, err := Shell_NotifyIcon(NIM_ADD, &data); err != nil {
		panic(err)
	}
}

func CreateMainWindow() (uintptr, error) {
	hInstance, err := GetModuleHandle(nil)
	if err != nil {
		return 0, err
	}
	wndClass := windows.StringToUTF16Ptr("DDNS")

	var wcex WNDCLASSEX
	wcex.CbSize = uint32(unsafe.Sizeof(wcex))
	wcex.LpfnWndProc = windows.NewCallback(wndProc)
	wcex.HInstance = hInstance
	wcex.LpszClassName = wndClass
	if _, err := RegisterClassEx(&wcex); err != nil {
		return 0, err
	}

	hwnd, err := CreateWindowEx(0, wndClass, windows.StringToUTF16Ptr("DDNS"), WS_OVERLAPPEDWINDOW, CW_USEDEFAULT, CW_USEDEFAULT, 400, 300, 0, 0, hInstance, nil)
	if err != nil {
		return 0, err
	}
	return hwnd, nil
}

func ShowWindow(hWnd uintptr, nCmdShow int32) (int32, error) {
	r, _, err := procShowWindow.Call(hWnd, uintptr(nCmdShow))
	if r == 0 {
		return 0, err
	}
	return int32(r), nil
}
