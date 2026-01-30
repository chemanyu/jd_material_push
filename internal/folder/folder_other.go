//go:build !windows

package folder

import "errors"

// SelectFolder 在非 Windows 平台返回错误
func SelectFolder() (string, error) {
	return "", errors.New("folder selection is only supported on Windows")
}
