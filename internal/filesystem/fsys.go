package filesystem

import (
	"io/fs"
)

type FS interface {
	Open(name string) (fs.File, error)
	fs.StatFS
	fs.ReadDirFS
	WriteFile(name string, data []byte, perm fs.FileMode) error
}
