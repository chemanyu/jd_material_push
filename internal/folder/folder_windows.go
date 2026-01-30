//go:build windows

package folder

import (
"syscall"
"unsafe"
)

var (
shell32                  = syscall.NewLazyDLL("shell32.dll")
procSHBrowseForFolder    = shell32.NewProc("SHBrowseForFolderW")
procSHGetPathFromIDList  = shell32.NewProc("SHGetPathFromIDListW")
)

const (
BIF_RETURNONLYFSDIRS = 0x00000001
BIF_NEWDIALOGSTYLE   = 0x00000040
BIF_USENEWUI         = BIF_NEWDIALOGSTYLE | 0x00000050
)

type BROWSEINFO struct {
	hwndOwner      uintptr
	pidlRoot       uintptr
	pszDisplayName *uint16
	lpszTitle      *uint16
	ulFlags        uint32
	lpfn           uintptr
	lParam         uintptr
	iImage         int32
}

// SelectFolder 显示 Windows 文件夹选择对话框
func SelectFolder() (string, error) {
	displayName := make([]uint16, syscall.MAX_PATH)
	title, _ := syscall.UTF16PtrFromString("请选择文件夹")
	
	bi := BROWSEINFO{
		hwndOwner:      0,
		pidlRoot:       0,
		pszDisplayName: &displayName[0],
		lpszTitle:      title,
		ulFlags:        BIF_RETURNONLYFSDIRS | BIF_NEWDIALOGSTYLE,
		lpfn:           0,
		lParam:         0,
		iImage:         0,
	}
	
	ret, _, _ := procSHBrowseForFolder.Call(uintptr(unsafe.Pointer(&bi)))
	if ret == 0 {
		return "", nil // 用户取消
	}
	
	path := make([]uint16, syscall.MAX_PATH)
	procSHGetPathFromIDList.Call(ret, uintptr(unsafe.Pointer(&path[0])))
	
	return syscall.UTF16ToString(path), nil
}
