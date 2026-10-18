//go:build windows

package app

import (
	"context"
	"fmt"
	"log"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	comdlg32             = syscall.NewLazyDLL("comdlg32.dll")
	procGetOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")

	user32                  = syscall.NewLazyDLL("user32.dll")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetActiveWindow     = user32.NewProc("GetActiveWindow")
)

// OPENFILENAMEW structure for GetOpenFileName
type openFileNameW struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        unsafe.Pointer
	dwReserved        uint32
	flagsEx           uint32
}

const (
	ofnFileMustExist   = 0x00001000
	ofnForceHide       = 0x00000010
	ofnNoChangeDir     = 0x00000008
)

func openFileDialog(_ context.Context) (string, error) {
	log.Printf("[Dialog] Opening file dialog via GetOpenFileName")

	// Build filter string: "All Files\0*.*\0\0"
	filter := utf16Ptr("所有文件\x00*.*\x00\x00")

	fileBuffer := make([]uint16, 1024)

	hwnd, _, _ := procGetActiveWindow.Call()
	log.Printf("[Dialog] Active window: 0x%x", hwnd)

	ofn := openFileNameW{
		lStructSize:  uint32(unsafe.Sizeof(openFileNameW{})),
		hwndOwner:    hwnd,
		lpstrFilter:  filter,
		nFilterIndex: 1,
		lpstrFile:    &fileBuffer[0],
		nMaxFile:     uint32(len(fileBuffer)),
		lpstrTitle:   utf16Ptr("选择要发送的文件"),
		flags:        ofnFileMustExist | ofnNoChangeDir,
	}

	log.Printf("[Dialog] Calling GetOpenFileName, struct size: %d", ofn.lStructSize)

	ret, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		commDlgErr, _, _ := comdlg32.NewProc("CommDlgExtendedError").Call()
		if commDlgErr != 0 {
			log.Printf("[Dialog] GetOpenFileName failed, extended error: 0x%x", commDlgErr)
			return "", fmt.Errorf("GetOpenFileName error: 0x%x", commDlgErr)
		}
		log.Printf("[Dialog] User cancelled file selection")
		return "", nil
	}

	result := utf16ToString(&fileBuffer[0])
	log.Printf("[Dialog] Selected: %s", result)
	return result, nil
}

func utf16Ptr(s string) *uint16 {
	u := utf16.Encode([]rune(s + "\x00"))
	return &u[0]
}

func utf16ToString(p *uint16) string {
	if p == nil {
		return ""
	}
	var n int
	q := p
	for *q != 0 {
		q = (*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(q)) + 2))
		n++
	}
	return string(utf16.Decode(unsafe.Slice(p, n)))
}
