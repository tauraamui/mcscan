//go:build !windows
// +build !windows

package vfs

const dotCharacter = 46

func IsHidden(path string) bool {
	if path[0] == dotCharacter {
		return true
	}

	return false
}
