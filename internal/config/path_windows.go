//go:build windows

// path_windows.go — the Windows Documents Known-Folder law: OneDrive-
// redirect aware via SHGetKnownFolderPath (FOLDERID_Documents), mirroring
// the reference's own ctypes approach (user_config.py:190-221) with the
// raw syscall instead of a dependency — Phase 1 stays stdlib-pure.
// golang.org/x/sys/windows.KnownFolderPath remains the doc-06-§2 option
// if a later phase prefers it. The Windows leg joins the Phase 7
// clean-machine checklist for live verification (compile-checked now via
// GOOS=windows).
package config

import (
	"os"
	"path/filepath"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

// folderIDDocuments is FOLDERID_Documents
// {FDD39AD0-238F-46AF-ADB4-6C85480369C7}.
type folderIDDocuments struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

var documentsGUID = folderIDDocuments{
	0xFDD39AD0, 0x238F, 0x46AF,
	[8]byte{0xAD, 0xB4, 0x6C, 0x85, 0x48, 0x03, 0x69, 0xC7},
}

var (
	shell32      = syscall.NewLazyDLL("shell32.dll")
	ole32        = syscall.NewLazyDLL("ole32.dll")
	getKnownPath = shell32.NewProc("SHGetKnownFolderPath")
	coTaskFree   = ole32.NewProc("CoTaskMemFree")
)

// windowsDocumentsDir mirrors _windows_documents_dir: the real Documents
// folder Windows itself calls Documents; "" on any failure (never
// raises).
func windowsDocumentsDir() string {
	var ptr *uint16
	ret, _, _ := getKnownPath.Call(
		uintptr(unsafe.Pointer(&documentsGUID)),
		0, // dwFlags
		0, // hToken (current user)
		uintptr(unsafe.Pointer(&ptr)),
	)
	if ret != 0 || ptr == nil {
		return ""
	}
	defer coTaskFree.Call(uintptr(unsafe.Pointer(ptr)))
	// decode the NUL-terminated UTF-16 string
	p := unsafe.Pointer(ptr)
	i := 0
	for *(*uint16)(unsafe.Add(p, 2*i)) != 0 {
		i++
	}
	units := make([]uint16, i)
	for j := 0; j < i; j++ {
		units[j] = *(*uint16)(unsafe.Add(p, 2*j))
	}
	return string(utf16.Decode(units)) // "" when empty
}

// documentsDir mirrors _documents_dir on Windows: the Known-Folder first,
// else ~/Documents only when it exists; "" when the feature sleeps.
func documentsDir() string {
	if known := windowsDocumentsDir(); known != "" {
		return known
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	cand := filepath.Join(home, "Documents")
	if fi, err := os.Stat(cand); err == nil && fi.IsDir() {
		return cand
	}
	return ""
}
