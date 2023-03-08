//go:build windows

package deamon

import (
	"fmt"
	"golang.org/x/sys/windows"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	IMAGE_ICON      = 1
	LR_DEFAULTSIZE  = 0x00000040
	LR_LOADFROMFILE = 0x00000010

	SW_HIDE          = 0
	SW_SHOWNORMAL    = 1
	SW_SHOWMINIMIZED = 2
	SW_SHOW          = 5

	CW_USEDEFAULT = ^0x7fffffff

	GWLP_WNDPROC = -4
	GCL_HMODULE  = -16

	TPM_BOTTOMALIGN = 0x0020
	TPM_LEFTALIGN   = 0x0000
	MF_STRING       = 0x00000000
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
	WM_CLOSE         = 0x0010
	WM_QUIT          = 0x0012
	WM_SHOWWINDOW    = 0x0018
	WM_SETICON       = 0x0080
	WM_COMMAND       = 0x111
	WM_MENUCHAR      = 0x120
	WM_MOUSEMOVE     = 0x0200
	WM_LBUTTONDOWN   = 0x0201
	WM_LBUTTONUP     = 0x0202
	WM_LBUTTONDBLCLK = 0x0203
	WM_RBUTTONDOWN   = 0x0204
	WM_RBUTTONUP     = 0x0205
	WM_RBUTTONDBLCLK = 0x0206
	WM_APP           = 0x8000

	TrayMsg = WM_APP + 1
)

const (
	NIM_ADD    = 0x00000000
	NIM_DELETE = 0x00000002

	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004
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

type WINDOWPLACEMENT struct {
	Length                       uint32
	Flags                        uint32
	ShowCmd                      uint32
	PtMinPosition, PtMaxPosition POINT
	RcNormalPosition, RcDevice   RECT
}

type POINT struct {
	X int32
	Y int32
}

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
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
	libKernel32 = windows.NewLazySystemDLL("kernel32.dll")
	libShell32  = windows.NewLazySystemDLL("shell32.dll")
	libUser32   = windows.NewLazySystemDLL("user32.dll")

	getConsoleWindow = libKernel32.NewProc("GetConsoleWindow")
	getModuleHandleW = libKernel32.NewProc("GetModuleHandleW")

	shellNotifyIcon = libShell32.NewProc("Shell_NotifyIconW")

	appendMenuW         = libUser32.NewProc("AppendMenuW")
	createPopupMenu     = libUser32.NewProc("CreatePopupMenu")
	createWindowExW     = libUser32.NewProc("CreateWindowExW")
	destroyMenu         = libUser32.NewProc("DestroyMenu")
	defWindowProcW      = libUser32.NewProc("DefWindowProcW")
	dispatchMessageW    = libUser32.NewProc("DispatchMessageW")
	getCursorPos        = libUser32.NewProc("GetCursorPos")
	getMessageW         = libUser32.NewProc("GetMessageW")
	getWindowPlacement  = libUser32.NewProc("GetWindowPlacement")
	loadImageW          = libUser32.NewProc("LoadImageW")
	postQuitMessage     = libUser32.NewProc("PostQuitMessage")
	registerClassExW    = libUser32.NewProc("RegisterClassExW")
	setWindowLong       = libUser32.NewProc("SetWindowLongW")
	showWindow          = libUser32.NewProc("ShowWindow")
	trackPopupMenu      = libUser32.NewProc("TrackPopupMenu")
	translateMessage    = libUser32.NewProc("TranslateMessage")
	setForegroundWindow = libUser32.NewProc("SetForegroundWindow")
)

func ShellNotifyIcon(dwMessage uint32, lpData *NOTIFYICONDATA) (int32, error) {
	r, _, err := shellNotifyIcon.Call(uintptr(dwMessage), uintptr(unsafe.Pointer(lpData)))
	if r == 0 {
		return 0, err
	}
	return int32(r), nil
}

func LoadImage(hInst uintptr, name *uint16, type_ uint32, cx, cy int32, fuLoad uint32) (uintptr, error) {
	r, _, err := loadImageW.Call(hInst, uintptr(unsafe.Pointer(name)), uintptr(type_), uintptr(cx), uintptr(cy), uintptr(fuLoad))
	if r == 0 {
		return 0, err
	}
	return r, nil
}

func GetConsoleWindow() (uintptr, error) {
	hWnd, _, err := getConsoleWindow.Call()
	if hWnd == 0 {
		return 0, err
	}
	return hWnd, nil
}

func GetModuleHandle(lpModuleName *uint16) (uintptr, error) {
	r, _, err := getModuleHandleW.Call(uintptr(unsafe.Pointer(lpModuleName)))
	if r == 0 {
		return 0, err
	}
	return r, nil
}

func RegisterClassEx(Arg1 *WNDCLASSEX) (uint16, error) {
	r, _, err := registerClassExW.Call(uintptr(unsafe.Pointer(Arg1)))
	if r == 0 {
		return 0, err
	}
	return uint16(r), nil
}

func CreateWindowEx(dwExStyle uint32, lpClassName, lpWindowName *uint16, dwStyle uint32, X, Y, nWidth, nHeight int32, hWndParent, HMenu, hInstance uintptr, lpParam unsafe.Pointer) (uintptr, error) {
	r, _, err := createWindowExW.Call(uintptr(dwExStyle), uintptr(unsafe.Pointer(lpClassName)), uintptr(unsafe.Pointer(lpWindowName)), uintptr(dwStyle), uintptr(X), uintptr(Y), uintptr(nWidth), uintptr(nHeight), hWndParent, HMenu, hInstance, uintptr(lpParam))
	if r == 0 {
		return 0, err
	}
	return r, nil
}

func DefWindowProc(hWnd uintptr, Msg uint32, wParam, lParam uintptr) (uintptr, error) {
	r, _, _ := defWindowProcW.Call(hWnd, uintptr(Msg), wParam, lParam)
	return r, nil
}

func DestroyMenu(hMenu uintptr) error {
	r, _, err := destroyMenu.Call(hMenu)
	if r == 0 {
		return err
	}
	return nil
}

func GetMessage(lpMsg *MSG, hWnd uintptr, uMsgFilterMin, uMsgFilterMax uint32) (int32, error) {
	r, _, err := getMessageW.Call(uintptr(unsafe.Pointer(lpMsg)), hWnd, uintptr(uMsgFilterMin), uintptr(uMsgFilterMax))
	if int32(r) == -1 {
		return 0, err
	}
	return int32(r), nil
}

func PostQuitMessage(nExitCode int32) {
	_, _, _ = postQuitMessage.Call(uintptr(nExitCode))
}

func TranslateMessage(lpMsg *MSG) (int32, error) {
	r, _, _ := translateMessage.Call(uintptr(unsafe.Pointer(lpMsg)))
	return int32(r), nil
}

func DispatchMessage(lpMsg *MSG) (uintptr, error) {
	r, _, _ := dispatchMessageW.Call(uintptr(unsafe.Pointer(lpMsg)))
	return r, nil
}

func SetWinProc(hWnd uintptr, proc func(uintptr, uint32, uintptr, uintptr) uintptr) (uintptr, error) {
	var offset int32 = GWLP_WNDPROC
	r, _, err := setWindowLong.Call(hWnd, uintptr(offset), windows.NewCallback(proc))
	if r == 0 {
		return 0, err
	}
	return r, nil
}

func CheckWindowMinimize(hWnd uintptr) (bool, error) {

	var lpwndpl WINDOWPLACEMENT

	r, _, err := getWindowPlacement.Call(hWnd, uintptr(unsafe.Pointer(&lpwndpl)))
	if r == 0 {
		return false, err
	}

	if lpwndpl.ShowCmd == SW_SHOWMINIMIZED {
		return true, nil
	}
	return false, nil
}

func CreateMainWindow(title string, callback func(uintptr, uint32, uintptr, uintptr) uintptr) (uintptr, error) {
	hInstance, err := GetModuleHandle(nil)
	if err != nil {
		return 0, err
	}
	wndClass := windows.StringToUTF16Ptr(title)

	var wcex WNDCLASSEX
	wcex.CbSize = uint32(unsafe.Sizeof(wcex))
	wcex.LpfnWndProc = windows.NewCallback(callback)
	wcex.HInstance = hInstance
	wcex.LpszClassName = wndClass
	if _, err := RegisterClassEx(&wcex); err != nil {
		return 0, err
	}

	hWnd, err := CreateWindowEx(0, wndClass, windows.StringToUTF16Ptr(title), WS_OVERLAPPEDWINDOW, CW_USEDEFAULT, CW_USEDEFAULT, 400, 300, 0, 0, hInstance, nil)
	if err != nil {
		return 0, err
	}
	return hWnd, nil
}

func ShowMenu(hWnd uintptr, hMenu uintptr) error {

	point := POINT{}
	r, _, err := getCursorPos.Call(uintptr(unsafe.Pointer(&point)))
	if r == 0 {
		return err
	}

	r, _, err = setForegroundWindow.Call(hWnd)

	r, _, err = trackPopupMenu.Call(hMenu, TPM_BOTTOMALIGN|TPM_LEFTALIGN, uintptr(point.X), uintptr(point.Y), 0, hWnd, 0)
	if r == 0 {
		return err
	}
	return nil
}

func ShowWindow(hWnd uintptr, nCmdShow int32) (int32, error) {
	r, _, err := showWindow.Call(hWnd, uintptr(nCmdShow))
	if r == 0 {
		return 0, err
	}
	return int32(r), nil
}

type GUI struct {
	consoleHWND uintptr
	trayHWND    uintptr
	trayData    NOTIFYICONDATA
	hMenu       uintptr

	show chan bool
	exit int32
}

func (g *GUI) Init() error {
	var err error
	g.consoleHWND, err = GetConsoleWindow()
	if err != nil {
		return err
	}

	g.trayHWND, err = CreateMainWindow(WindowName, g.wndProc)
	if err != nil {
		return err
	}
	_, _ = ShowWindow(g.trayHWND, SW_HIDE)

	if err = g.initTrayData(); err != nil {
		return err
	}

	if err = g.createMenu(); err != nil {
		return err
	}

	if _, err := ShellNotifyIcon(NIM_ADD, &g.trayData); err != nil {
		return err
	}

	atomic.StoreInt32(&g.exit, 0)
	g.show = make(chan bool, 10)
	return nil
}

func (g *GUI) initTrayData() error {
	icon, err := LoadImage(0, windows.StringToUTF16Ptr(ICONPath), IMAGE_ICON, 0, 0, LR_DEFAULTSIZE|LR_LOADFROMFILE)
	if err != nil {
		return err
	}

	g.trayData = NOTIFYICONDATA{}
	g.trayData.CbSize = uint32(unsafe.Sizeof(g.trayData))
	g.trayData.UFlags = NIF_ICON | NIF_TIP | NIF_MESSAGE
	g.trayData.UCallbackMessage = TrayMsg
	g.trayData.HWnd = g.trayHWND
	g.trayData.HIcon = icon
	copy(g.trayData.SzTip[:], windows.StringToUTF16(WindowName))
	return nil
}

func (g *GUI) createMenu() error {
	hMenu, _, err := createPopupMenu.Call()
	if hMenu == 0 {
		return err
	}
	r, _, err := appendMenuW.Call(hMenu, MF_STRING, WM_CLOSE, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("退出"))))
	if r == 0 {
		return err
	}
	g.hMenu = hMenu
	return nil
}

func (g *GUI) Run(fn func(chan bool)) {
	defer func() {
		if _, err := ShellNotifyIcon(NIM_DELETE, &g.trayData); err != nil {
		}
		close(g.show)
	}()

	//SetWinProc(gui.consoleHWND, wndProc)

	var wg = sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		//check window show
		var ticker = time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for atomic.LoadInt32(&g.exit) == 0 {
			select {
			case show, ok := <-g.show:
				if !ok {
					return
				}
				if show {
					ticker.Reset(500 * time.Millisecond)
				} else {
					ticker.Stop()
				}
			case <-ticker.C:
				//
			}

			r, err := CheckWindowMinimize(g.consoleHWND)
			if err != nil {
				continue
			}
			if r {
				_, _ = ShowWindow(g.consoleHWND, SW_HIDE)
				ticker.Stop()
			}
		}
	}()

	if _, err := ShowWindow(g.consoleHWND, SW_SHOW); err != nil {
		fmt.Println(err)
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var msg MSG
		for atomic.LoadInt32(&g.exit) == 0 {
			r, err := GetMessage(&msg, 0, 0, 0)
			if r == 0 || err != nil {
				fmt.Println(err)
				break
			}

			_, _ = TranslateMessage(&msg)
			_, _ = DispatchMessage(&msg)
		}
	}()

	workExit := make(chan bool, 1)
	defer close(workExit)
	go fn(workExit)
	wg.Wait()
	workExit <- true
}

func (g *GUI) stop() {
	if atomic.CompareAndSwapInt32(&g.exit, 0, 1) {
		g.show <- true
		PostQuitMessage(0)
		_, _, _ = destroyMenu.Call(g.hMenu)
		_, _ = ShellNotifyIcon(NIM_DELETE, &g.trayData)
	}
}

func (g *GUI) wndProc(hWnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case TrayMsg:
		switch nMsg := uint16(lParam); nMsg {
		case WM_LBUTTONDOWN:
			_, _ = ShowWindow(g.consoleHWND, SW_SHOWNORMAL)
			g.show <- true
		case WM_RBUTTONDOWN:
			_ = ShowMenu(g.trayHWND, g.hMenu)
		}
	case WM_SHOWWINDOW:
	//TODO
	case WM_MENUCHAR:
		//fmt.Println(wParam, lParam)
	case WM_COMMAND:
		switch nMsg := uint16(wParam); nMsg {
		case WM_CLOSE:
			g.stop()
		}
	case WM_CLOSE | WM_DESTROY:
		g.stop()
	default:
		r, _ := DefWindowProc(hWnd, msg, wParam, lParam)
		return r
	}
	return 0
}
